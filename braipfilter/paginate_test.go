package braipfilter

import (
	"reflect"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
)

func TestDBFilter_paginateConfig(t *testing.T) {
	type sample struct {
		Page    int
		PerPage int
		OrderBy []string
	}

	type fields struct {
		defaultValues defaultValues
	}
	type args struct {
		filterStruct interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   paginateConfig
	}{
		{
			name: "all ok",
			args: args{
				filterStruct: sample{
					Page:    2,
					PerPage: 123,
					OrderBy: []string{"id.asc"},
				},
			},
			want: paginateConfig{
				page:    2,
				perPage: 123,
				orderBy: "id asc",
			},
		},
		{
			name: "default values",
			fields: fields{
				defaultValues: defaultValues{
					perPage: perPageDefault,
				},
			},
			args: args{
				filterStruct: sample{
					Page:    0,
					PerPage: 0,
				},
			},
			want: paginateConfig{
				page:    1,
				perPage: perPageDefault,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &DBFilter{
				defaultValues: tt.fields.defaultValues,
			}
			if got := f.paginateConfig(tt.args.filterStruct); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBFilter.paginateConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_paginateConfig_offset(t *testing.T) {
	type fields struct {
		page    int
		perPage int
		orderBy string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "success case",
			fields: fields{
				page:    2,
				perPage: 11,
			},
			want: 11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &paginateConfig{
				page:    tt.fields.page,
				perPage: tt.fields.perPage,
				orderBy: tt.fields.orderBy,
			}
			if got := c.offset(); got != tt.want {
				t.Errorf("paginateConfig.offset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_totalPages(t *testing.T) {
	type args struct {
		count   int
		perPage int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "success case",
			args: args{
				count:   101,
				perPage: 10,
			},
			want: 11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := totalPages(tt.args.count, tt.args.perPage); got != tt.want {
				t.Errorf("totalPages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaginate_fill(t *testing.T) {
	type fields struct {
		Total        int
		Page         int
		PerPage      int
		NextPage     *int
		PreviousPage *int
		TotalPages   int
		To           int
	}
	type args struct {
		page          int
		perPage       int
		offset        int
		count         int
		filteredCount int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *PaginatePageBased
	}{
		{
			name: "success all cases",
			args: args{
				page:          2,
				perPage:       10,
				offset:        10,
				count:         123,
				filteredCount: 10,
			},
			want: &PaginatePageBased{
				Total:        123,
				Page:         2,
				PerPage:      10,
				NextPage:     pointer.ToInt(3),
				PreviousPage: pointer.ToInt(1),
				TotalPages:   13,
				To:           20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PaginatePageBased{
				Total:        tt.fields.Total,
				Page:         tt.fields.Page,
				PerPage:      tt.fields.PerPage,
				NextPage:     tt.fields.NextPage,
				PreviousPage: tt.fields.PreviousPage,
				TotalPages:   tt.fields.TotalPages,
				To:           tt.fields.To,
			}
			p.fill(tt.args.page, tt.args.perPage, tt.args.offset, tt.args.count, tt.args.filteredCount)

			assert.Equal(t, tt.want, p)
		})
	}
}
