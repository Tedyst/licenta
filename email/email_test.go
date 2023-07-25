package email

import "testing"

func TestSendMultipartEmail(t *testing.T) {
	type args struct {
		subject string
		address string
		html    string
		text    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestSendMultipartEmail",
			args: args{
				subject: "test",
				address: "test",
				html:    "test",
				text:    "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMultipartEmailDebug(tt.args.address, tt.args.subject, tt.args.html, tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("SendMultipartEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
