<% template "admin-header.html" .%>

<span class="pull-right"><a class="btn btn-primary" href="/admin/users/edit" title="Add new user"><i class="icon-plus-sign icon-white"></i> Add New</a></span>
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
      <td><a href="/admin/users/edit/id/<% $user.Id %>"><% $user.Username %></a></td>
      <td><% $user.FullName %></td>
      <td><% $user.Email %></td>
      <td><% dateformat $user.CreatedOn "2 Jan 2006"%></td> 
    </tr>
    <%end%>
  </tbody>
</table>

<% template "admin-footer.html" .%>