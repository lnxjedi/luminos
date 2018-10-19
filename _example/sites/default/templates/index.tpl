
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

      {{ if .Content }}
        {{ if .TOC }}<h1>Contents:</h1>{{ end }}

        {{ .Content }}

      {{ else }}

        {{ if .CurrentPage }}
          <h1>{{ .CurrentPage.Text }}</h1>
        {{ end }}

        <ul>
          {{ range .SideMenu }}
            <li>
              <a href="{{ asset .URL }}">{{ .Text }}</a>
            </li>
          {{ end }}
        </ul>

      {{end}}

      {{ if .Site.page.body.copyright }}
        <p>{{ .Site.page.body.copyright | html }}</p>
      {{ end }}

    </div>

  {{ if .Site.page.body.scripts.footer }}
    <script type="text/javascript">
      {{ .Settigns.page.body.scripts.footer | js }}
    </script>
  {{ end }}

  </body>
</html>
