<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en-us">

  <head>
    <link href="http://gmpg.org/xfn/11" rel="profile">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta http-equiv="content-type" content="text/html; charset=utf-8">

    <!-- Enable responsiveness on mobile devices-->
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1">

    {{ $title := "Luminos Website" }}
    {{ if .Site.Page.Head.Title }}
      {{ $title = .Site.Page.Head.Title }}
    {{ end }}

    <title>
      {{ if .IsHome }}
        {{ .Site.Page.Head.Title }}
      {{ else }}
        {{ if .Title }}
          {{ .Title }} {{ if .Site.Page.Head.Title }} &middot; {{ $title }} {{ end }}
        {{ else }}
          {{ $title }}
        {{ end }}
      {{ end }}
    </title>

    <!-- CSS -->
    <link rel="stylesheet" href="{{ asset "/css/poole.css" }}">
    <link rel="stylesheet" href="{{ asset "/css/syntax.css" }}">
    <link rel="stylesheet" href="{{ asset "/css/hyde.css" }}">

    <!-- External fonts -->
    <link href="//fonts.googleapis.com/css?family=Source+Code+Pro" rel="stylesheet" type="text/css">
    <link href="//fonts.googleapis.com/css?family=Open+Sans:300,400italic,400,600,700|Abril+Fatface" rel="stylesheet" type="text/css">

    <!-- Icons -->
    <link rel="apple-touch-icon-precomposed" sizes="144x144" href="{{ asset "/apple-touch-icon-precomposed.png" }}">
    <link rel="shortcut icon" href="{{ asset "/favicon.ico"}}">

    <!-- Code highlighting -->
    <link rel="stylesheet" href="//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.12.0/styles/atom-one-light.min.css">
    <script src="//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.12.0/highlight.min.js"></script>
    <script>hljs.initHighlightingOnLoad();</script>

    <!-- Luminos -->
    <link rel="stylesheet" href="{{ asset "/css/luminos.css" }}">
    <script src="{{asset "js/main.js"}}"></script>

  </head>
