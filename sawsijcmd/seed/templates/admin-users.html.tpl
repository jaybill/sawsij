<% template "admin-header.html" .%>

<h1>Manage Users</h1>

<table class="table table-striped table-bordered table-condensed">
  <thead>
    <tr>
      <th>Username</th>
      <th>Full Name</th>
      <th>Email</th>
      <th>Created On</th>
    </tr>
  </thead>
  <tbody>
    <%range $index,$user := .users%>
    <tr>
      <td><a href="/admin/user/edit/id/<% $user.Id %>"><% $user.Username %></a></td>
      <td><% $user.FullName %></td>
      <td><% $user.Email %></td>
      <td><% dateformat $user.CreatedOn "2 Jan 2006"%></td> 
    </tr>
    <%end%>
  </tbody>
</table>

<% template "admin-footer.html" .%>