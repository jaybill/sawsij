<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="">
    <meta name="author" content="">

    <title>{{.name}}</title>

    <!-- Bootstrap core CSS -->
    <link rel="stylesheet" href="//netdna.bootstrapcdn.com/bootstrap/3.0.0-wip/css/bootstrap.min.css">

    <!-- Custom styles for this template -->
    <link href="/static/css/site.css" rel="stylesheet">
  </head>

  <body>

  <div class="navbar navbar-default navbar-fixed-top">
    <div class="navbar-header">
      <a class="navbar-brand" href="/">{{.name}}</a>      
    </div>
      <ul class="nav navbar-nav navbar-right">
      <% if .global.user %>  
        <% if equal .global.user.Role .global.roles.admin %>  
        <li><a href="/admin">Admin</a></li>
        <% end %>              
      <li><p class="navbar-text">Logged in as <strong><% .global.user.Username %></strong></p></li>
      <li><a href="/logout">Log Out</a></li> 
      <% else %>
      <li><a href="/login">Log In</a></li>
      <% end %>
      </ul>
  </div>
  <div class="container">
  <% template "messages.html" .%>