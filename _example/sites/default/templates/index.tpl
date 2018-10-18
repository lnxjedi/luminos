
{{ template "header.tpl" . }}

  <body>

    <!-- sidebar -->
    <div class="sidebar">

      <div class="container">

        {{ if settings "page/body/menu_pull" }}
          <ul class="nav nav-tabs">
          {{ range settings "page/body/menu_pull" }}
            <li><a href="{{ asset .URL }}">{{ .Text }}</a></li>
          {{ end }}
          </ul>
        {{ end }}

        <div class="sidebar-about">
          <div class="logo">
            <a href="/">
              <!--
              Icon made by OCHA (http://www.unocha.org) from www.flaticon.com
              is licensed under CC BY 3.0
              (http://creativecommons.org/licenses/by/3.0/)
              -->
              <img src="{{ asset "/images/logo.svg" }}" width="128" height="128" />
            </a>
          </div>
          <h1>
            <a href="{{ asset "/" }}">
              {{ setting "page/brand" }}
            </a>
          </h1>
          <p class="lead">{{ setting "page/body/title" }}</p>
        </div>

        <nav class="sidebar-nav">
          {{ if .IsHome }}
            {{ range settings "page/body/menu" }}
              <a class="sidebar-nav-item" href="{{ asset .URL }}">{{ .Text }}</a>
            {{ end }}

          {{ else }}
            {{ if .SideMenu }}
              {{ range .SideMenu }}
                <a class="sidebar-nav-item" href="{{ asset .URL }}">{{ .Text }}</a>
              {{ end }}
            {{ else }}
              {{ range settings "page/body/menu" }}
                <a class="sidebar-nav-item" href="{{ asset .URL }}">{{ .Text }}</a>
              {{ end }}
            {{ end }}
          {{ end }}
        </nav>

      </div>

      {{ if not .IsHome }}
        {{ if .SideMenu }}
          {{ if settings "page/body/menu" }}
            <div class="collapse navbar-collapse">
              <ul class="nav navbar-nav">
              {{ range settings "page/body/menu" }}
                <li><a href="{{ asset .URL }}">{{ .Text }}</a></li>
              {{ end }}
              </ul>
            </div>
          {{ end }}
        {{ end }}
      {{ end }}

    </div>

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

      {{ if setting "page/body/copyright" }}
        <p>{{ setting "page/body/copyright" | html }}</p>
      {{ end }}

    </div>

  {{ if setting "page/body/scripts/footer" }}
    <script type="text/javascript">
      {{ setting "page/body/scripts/footer" | js }}
    </script>
  {{ end }}

  </body>
</html>
