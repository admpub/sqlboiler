{{- define "timestamp_insert_helper" -}}
	{{- if not .NoAutoTimestamps -}}
	{{- $colNames := .Table.Columns | columnNames -}}
	{{if containsAny $colNames "created_at" "updated_at"}}
	currTime := time.Now().In(boil.GetLocation())
		{{range $ind, $col := .Table.Columns}}
			{{- if eq $col.Name "created_at" -}}
				{{- if eq $col.Type "time.Time" }}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}
				{{- else if eq $col.Type "uint" }}
	if o.CreatedAt == 0 {
		o.CreatedAt = uint(currTime.Unix())
	}
				{{- else if eq $col.Type "uint64" }}
	if o.CreatedAt == 0 {
		o.CreatedAt = uint64(currTime.Unix())
	}
				{{- else}}
	if queries.MustTime(o.CreatedAt).IsZero() {
		queries.SetScanner(&o.CreatedAt, currTime)
	}
				{{- end -}}
			{{- end -}}
			{{- if eq $col.Name "updated_at" -}}
				{{- if eq $col.Type "time.Time"}}
	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = currTime
	}
				{{- else if eq $col.Type "uint" }}
	if o.UpdatedAt == 0 {
		o.UpdatedAt = uint(currTime.Unix())
	}
				{{- else if eq $col.Type "uint64" }}
	if o.UpdatedAt == 0 {
		o.UpdatedAt = uint64(currTime.Unix())
	}
				{{- else}}
	if queries.MustTime(o.UpdatedAt).IsZero() {
		queries.SetScanner(&o.UpdatedAt, currTime)
	}
				{{- end -}}
			{{- end -}}
		{{end}}
	{{end}}
	{{- end}}
{{- end -}}
{{- define "timestamp_update_helper" -}}
	{{- if not .NoAutoTimestamps -}}
	{{- $colNames := .Table.Columns | columnNames -}}
	{{if containsAny $colNames "updated_at"}}
	currTime := time.Now().In(boil.GetLocation())
		{{range $ind, $col := .Table.Columns}}
			{{- if eq $col.Name "updated_at" -}}
				{{- if eq $col.Type "time.Time"}}
	o.UpdatedAt = currTime
				{{- else if eq $col.Type "uint" }}
	o.UpdatedAt = uint(currTime.Unix())
				{{- else if eq $col.Type "uint64" }}
	o.UpdatedAt = uint64(currTime.Unix())
				{{- else}}
	queries.SetScanner(&o.UpdatedAt, currTime)
				{{- end -}}
			{{- end -}}
		{{end}}
	{{end}}
	{{- end}}
{{end -}}
{{- define "timestamp_upsert_helper" -}}
	{{- if not .NoAutoTimestamps -}}
	{{- $colNames := .Table.Columns | columnNames -}}
	{{if containsAny $colNames "created_at" "updated_at"}}
	currTime := time.Now().In(boil.GetLocation())
		{{range $ind, $col := .Table.Columns}}
			{{- if eq $col.Name "created_at" -}}
				{{- if eq $col.Type "time.Time"}}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}
				{{- else if eq $col.Type "uint" }}
	if o.CreatedAt == 0 {
		o.CreatedAt = uint(currTime.Unix())
	}
				{{- else if eq $col.Type "uint64" }}
	if o.CreatedAt == 0 {
		o.CreatedAt = uint64(currTime.Unix())
	}
				{{- else}}
	if queries.MustTime(o.CreatedAt).IsZero() {
		queries.SetScanner(&o.CreatedAt, currTime)
	}
				{{- end -}}
			{{- end -}}
			{{- if eq $col.Name "updated_at" -}}
				{{- if eq $col.Type "time.Time"}}
	o.UpdatedAt = currTime
				{{- else if eq $col.Type "uint" }}
	o.UpdatedAt = uint(currTime.Unix())
				{{- else if eq $col.Type "uint64" }}
	o.UpdatedAt = uint64(currTime.Unix())
				{{- else}}
	queries.SetScanner(&o.UpdatedAt, currTime)
				{{- end -}}
			{{- end -}}
		{{end}}
	{{end}}
	{{- end}}
{{end -}}
