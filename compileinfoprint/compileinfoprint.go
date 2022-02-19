// compileinfoprint is imported for the side effect of printing the compileinfo
// to os.StdErr
package compileinfoprint

import "github.com/carbocation/genomisc/compileinfo"

func init() {
	compileinfo.PrintToStdErr()
}
