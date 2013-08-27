package resources

func GetStaticResources() (r map[string]string) {

	r = map[string]
	string{
		{{ range $s := .static }}"{{ $s.Name }}":  "{{ $s.Content }}",
		{{ end }}
	}
	return

}
