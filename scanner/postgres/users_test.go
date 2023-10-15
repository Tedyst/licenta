package postgres

import (
	"testing"
)

func Test_postgresUser_VerifyPassword(t *testing.T) {
	type fields struct {
		super    bool
		name     string
		password string
	}
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			args: args{
				password: "postgres",
			},
			fields: fields{
				super:    false,
				name:     "postgres",
				password: "SCRAM-SHA-256$4096:x1Y7a1TyFE4fFwUOMyvX8Q==$WvMDOS/ZDzXzaHjpPqzkqXrd1ntcDIi7P2jQwYgI0e4=:Wf431GYj+SeayVQ6zOijoV5xQwzZKyGoVU7IAbXTO7U=",
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				password: "postgres123",
			},
			fields: fields{
				super:    false,
				name:     "postgres",
				password: "SCRAM-SHA-256$4096:x1Y7a1TyFE4fFwUOMyvX8Q==$WvMDOS/ZDzXzaHjpPqzkqXrd1ntcDIi7P2jQwYgI0e4=:Wf431GYj+SeayVQ6zOijoV5xQwzZKyGoVU7IAbXTO7U=",
			},
			want:    false,
			wantErr: false,
		},
		{
			args: args{
				password: "postgres",
			},
			fields: fields{
				super:    false,
				name:     "postgres123",
				password: "SCRAM-SHA-256$4096:x1Y7a1TyFE4fFwUOMyvX8Q==$WvMDOS/ZDzXzaHjpPqzkqXrd1ntcDIi7P2jQwYgI0e4=:Wf431GYj+SeayVQ6zOijoV5xQwzZKyGoVU7IAbXTO7U=",
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				password: "postgres",
			},
			fields: fields{
				super:    false,
				name:     "postgres",
				password: "SCRAM-SHA-255$asdasdassad",
			},
			want:    false,
			wantErr: false,
		},
		{
			args: args{
				password: "postgres",
			},
			fields: fields{
				super:    false,
				name:     "postgres",
				password: "postgres",
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				password: "postgres",
			},
			fields: fields{
				super:    false,
				name:     "foo1",
				password: "md5b4fe08fbb9d193ffd48c3a10cbf2a04c",
			},
			want:    false,
			wantErr: false,
		},
		{
			args: args{
				password: "postgres",
			},
			fields: fields{
				super:    false,
				name:     "foo2",
				password: "md5b4fe08fbb9d193ffd48c3a10cbf2a04c",
			},
			want:    false,
			wantErr: false,
		},
		{
			args: args{
				password: "secret",
			},
			fields: fields{
				super:    false,
				name:     "foo1",
				password: "md5b4fe08fbb9d193ffd48c3a10cbf2a04c",
			},
			want:    true,
			wantErr: false,
		},
		{
			args: args{
				password: "secret",
			},
			fields: fields{
				super:    false,
				name:     "foo2",
				password: "md5b4fe08fbb9d193ffd48c3a10cbf2a04c",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &postgresUser{
				super:    tt.fields.super,
				name:     tt.fields.name,
				password: tt.fields.password,
			}
			got, err := u.VerifyPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("postgresUser.VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("postgresUser.VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
