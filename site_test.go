package crawler

import (
	"reflect"
	"sync"
	"testing"

	"github.com/ariefrahmansyah/href"
)

func TestSite_AppendSite(t *testing.T) {
	type args struct {
		s Site
	}
	tests := []struct {
		name           string
		site           *Site
		args           args
		wantTotalSites int
	}{
		{
			"0 -> 1",
			&Site{
				mutex: &sync.Mutex{},
			},
			args{
				Site{},
			},
			1,
		},
		{
			"1 -> 2",
			&Site{
				mutex: &sync.Mutex{},
				Sites: []Site{
					Site{},
				},
			},
			args{
				Site{},
			},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.site.AppendSite(tt.args.s)

			if len(tt.site.Sites) != tt.wantTotalSites {
				t.Errorf("len(site) = %v, want %v", len(tt.site.Sites), tt.wantTotalSites)
			}
		})
	}
}

func TestSortSites(t *testing.T) {
	type args struct {
		sites []Site
	}
	tests := []struct {
		name string
		args args
		want []Site
	}{
		{
			"empty site",
			args{
				[]Site{},
			},
			[]Site{},
		},
		{
			"one site",
			args{
				[]Site{
					Site{
						Data: href.Link{Text: "Site 1"},
					},
				},
			},
			[]Site{
				Site{
					Data: href.Link{Text: "Site 1"},
				},
			},
		},
		{
			"two sites, descending",
			args{
				[]Site{
					Site{
						Data: href.Link{Text: "Site 2"},
					},
					Site{
						Data: href.Link{Text: "Site 1"},
					},
				},
			},
			[]Site{
				Site{
					Data: href.Link{Text: "Site 1"},
				},
				Site{
					Data: href.Link{Text: "Site 2"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortSites(tt.args.sites)

			if !reflect.DeepEqual(tt.args.sites, tt.want) {
				t.Errorf("SortSites() = %v, want %v", tt.args.sites, tt.want)
			}
		})
	}
}
