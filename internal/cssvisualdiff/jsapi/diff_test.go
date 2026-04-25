package jsapi

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestDiffReportAcceptsCamelCaseChangeCount(t *testing.T) {
	vm := goja.New()
	exports := vm.NewObject()
	installDiffAPI(nil, vm, exports)
	require.NoError(t, vm.Set("cvd", exports))

	value, err := vm.RunString(`(() => {
  const diff = cvd.diff({ value: 1 }, { value: 2 });
  return cvd.report(diff).markdown();
})()`)
	require.NoError(t, err)
	require.Contains(t, value.String(), "1 change(s)")
	require.Contains(t, value.String(), "value")
}
