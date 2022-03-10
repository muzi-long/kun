package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/RussellLuo/kun/gen"
	"github.com/RussellLuo/kun/gen/util/annotation"
)

type userFlags struct {
	outDir        string
	flatLayout    bool
	testFileName  string
	formatted     bool
	snakeCase     bool
	enableTracing bool
	force         bool

	args []string
}

func main() {
	var flags userFlags
	flag.StringVar(&flags.outDir, "out", ".", "output directory")
	flag.BoolVar(&flags.flatLayout, "flat", true, "whether to use flat layout")
	flag.StringVar(&flags.testFileName, "test", "./http.test.yaml", "the YAML file that provides test-cases for HTTP")
	flag.BoolVar(&flags.formatted, "fmt", true, "whether to make code formatted")
	flag.BoolVar(&flags.snakeCase, "snake", true, "whether to use snake-case for default names")
	flag.BoolVar(&flags.enableTracing, "trace", false, "whether to enable tracing")
	flag.BoolVar(&flags.force, "force", false, "whether to remove previously generated files before generating new ones")

	flag.Usage = func() {
		fmt.Println(`kungen [flags] source-file interface-name`)
		flag.PrintDefaults()
	}

	flag.Parse()
	flags.args = flag.Args()

	if err := run(flags); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(flags userFlags) error {
	if len(flags.args) != 2 {
		return errors.New("need 2 arguments")
	}

	srcFilename, interfaceName := flags.args[0], flags.args[1]

	srcFilename, err := filepath.Abs(srcFilename)
	if err != nil {
		return err
	}

	if flags.force {
		if err := removeGeneratedFiles(filepath.Dir(srcFilename)); err != nil {
			return err
		}
	}

	generator := gen.New(&gen.Options{
		OutDir:        flags.outDir,
		FlatLayout:    flags.flatLayout,
		SchemaPtr:     true,
		SchemaTag:     "json",
		SnakeCase:     flags.snakeCase,
		Formatted:     flags.formatted,
		EnableTracing: flags.enableTracing,
	})
	files, err := generator.Generate(srcFilename, interfaceName, flags.testFileName)
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := f.Write(); err != nil {
			return err
		}
	}

	return nil
}

// removeGeneratedFiles recursively remove all files generated by kun from dir.
func removeGeneratedFiles(dir string) error {
	isGenerated := func(path string) (bool, error) {
		f, err := os.Open(path)
		if err != nil {
			return false, err
		}
		defer f.Close()

		header := make([]byte, len(annotation.FileHeader))
		if _, err := io.ReadFull(f, header); err != nil {
			return false, err
		}

		return string(header) == annotation.FileHeader, nil
	}

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			// Ignore non-Go files.
			return nil
		}

		ok, err := isGenerated(path)
		if err != nil {
			return err
		}

		if ok {
			return os.Remove(path)
		}
		return nil
	})
}
