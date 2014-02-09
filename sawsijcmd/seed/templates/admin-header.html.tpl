<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="">
    <meta name="author" content="">

    <title>{{.name}} Admin</title>

    <!-- Bootstrap core CSS -->
     <link rel="stylesheet" href="//netdna.bootstrapcdn.com/bootstrap/3.0.0-wip/css/bootstrap.min.css">

    <!-- Custom styles for this template -->
    <link href="/static/css/admin.css" rel="stylesheet">
    <link href="/static/css/datepicker.css" rel="stylesheet">
  </head>

  <body>

  <div class="navbar navbar-default navbar-fixed-top">
    <div class="navbar-header">
      <a class="navbar-brand" href="/admin">{{.name}} Admin</a>      
    </div>
    <ul class="nav navbar-nav">      
      <li <% if equal .global.url "/admin" %>class="active"<% end %>><a href="/admin">Dashboard</a></li>   
      <li <% if equal .global.url "/admin/users" %>class="active"<% end %>><a href="/admin/users">Users</a></li>
    </ul>
    <ul class="nav navbar-nav navbar-right"> 
      <li><p class="navbar-text">Logged in as <strong><% .global.user.Username %></strong></p></li>
      <li><a href="/logout">Log Out</a></li>     
    </ul>
  </div>
  <div class="container">
  <% template "messages.html" .%>