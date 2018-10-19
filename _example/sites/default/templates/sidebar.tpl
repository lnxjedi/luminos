    <!-- sidebar -->
    <div class="sidebar">

      <div class="container">

        {{ if .Site.Page.Body.MenuPull }}
          <ul class="nav nav-tabs">
          {{ range .Site.Page.Body.MenuPull }}
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
              {{ .Site.Page.Brand }}
            </a>
          </h1>
          <p class="lead">{{ .Site.Page.Body.Title }}</p>
          <form action="/search" method="GET">
            <input type="text" name="terms"><br>
            <input type="submit" value="search">
          </form>
        </div>

        <nav class="sidebar-nav">
          {{ if .IsHome }}
            {{ range .Site.Page.Body.Menu }}
              <a class="sidebar-nav-item" href="{{ asset .URL }}">{{ .Text }}</a>
            {{ end }}

          {{ else }}
            {{ if .SideMenu }}
              {{ range .SideMenu }}
                <a class="sidebar-nav-item" href="{{ asset .URL }}">{{ .Text }}</a>
              {{ end }}
            {{ else }}
              {{ range .Site.Page.Body.Menu }}
                <a class="sidebar-nav-item" href="{{ asset .URL }}">{{ .Text }}</a>
              {{ end }}
            {{ end }}
          {{ end }}
        </nav>

      </div>

      {{ if not .IsHome }}
        {{ if .SideMenu }}
          {{ if .Site.Page.Body.Menu }}
            <div class="collapse navbar-collapse">
              <ul class="nav navbar-nav">
              {{ range .Site.Page.Body.Menu }}
                <li><a href="{{ asset .URL }}">{{ .Text }}</a></li>
              {{ end }}
              </ul>
            </div>
          {{ end }}
        {{ end }}
      {{ end }}

    </div>
