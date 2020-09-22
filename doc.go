//Package ifchanged is a collection of functions to perform callbacks in case of file changes (using sha256 hash)
//and / or missing files.
//
//Use case: Generating `css` file only if `scss` file has changed or `css` file is missing.
//
//Example:
//
//		err = ifchanged.NewIf().
//				Changed(fileName, fileName+".sha256").
//				Missing("somefile.txt").
//				Execute(func() error {
//						fmt.Printf("This has been called because \"somefile.txt\" is missing or %v has changed\n", fileName)
//						return nil
//				})
package ifchanged
