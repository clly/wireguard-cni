package ipam

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewIPAM(t *testing.T) {
	const parentPrefix = "10.0.0.0/8"
	testcases := map[string]struct {
		withDataDir bool
		prefixs     []string
	}{
		"EmptyDataDir": {
			withDataDir: false,
		},
		"WithDataDir": {
			withDataDir: true,
		},
		"WithStoredAllocs": {
			withDataDir: true,
			prefixs:     []string{"10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"},
		},
	}
	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			ctx := context.Background()
			d := t.TempDir()
			ipam, err := New(context.Background(), d, parentPrefix)
			r.NoError(err)
			for _, s := range testcase.prefixs {
				_, err := ipam.Ipamer.AcquireSpecificChildPrefix(ctx, parentPrefix, s)
				r.NoError(err)
			}
			r.NoError(ipam.Save(ctx))
			ipam, err = New(context.Background(), d, "10.0.0.0/8")
			r.NoError(err)
			r.NoError(ipam.Save(ctx))
			if !testcase.withDataDir {
				r.NoFileExists(IpamDataFile)
			} else {
				r.FileExists(filepath.Join(d, IpamDataFile))
			}

			if testcase.prefixs != nil {
				actual, err := ipam.Ipamer.ReadAllPrefixCidrs(ctx)

				r.NoError(err)
				sort.Strings(actual)
				expectedPrefixes := append(testcase.prefixs, parentPrefix)
				sort.Strings(expectedPrefixes)
				r.EqualValues(expectedPrefixes, actual)
			}

		})
	}
}
