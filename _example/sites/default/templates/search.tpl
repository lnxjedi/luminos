
{{ template "header.tpl" . }}

  <body>

  {{ template "sidebar.tpl" . }}

    <div class="content container">
      {{ if not .IsHome }}
        {{ if .BreadCrumb }}
          <ul class="breadcrumb">
            {{ range .BreadCrumb }}
              <li><a href="{{ asset .URL }}">{{ .Text }}</a></li>
            {{ end }}
          </ul>
        {{ end }}
      {{ end }}

      <!-- search -->
      {{ if .Query.terms }}
        {{ $res := .Search .Query.terms 50 }}
        {{ $nres := len $res }}
        {{ if eq $nres 0 }}
          <h3>No results</h3>
        {{ else }}
          <h3>Results:</h3>
          <ul>
          {{ range $res }}
            {{ $res := printf "%s" .StoreValue }}
            <li><a href="{{ $res }}">{{ $res }}</a></li>
          {{ end }}
          </ul>
        {{ end }}
      {{ else }}
      <h3>No results</h3>
      {{ end }}

      {{ if .Site.Page.Body.Copyright }}
        <p>{{ .Site.Page.Body.Copyright | html }}</p>
      {{ end }}

    </div>

  {{ if .Site.Page.Body.Scripts.Footer }}
    <script type="text/javascript">
      {{ .Site.Page.Body.Scripts.Footer | js }}
    </script>
  {{ end }}

  </body>
</html>
