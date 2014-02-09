<% template "admin-header.html" .%>

<span class="pull-right"><a href="/admin/users">Back to list &raquo;</a></span>
<h1>Manage Users</h1>
<h3><% if .update %>Edit User<% else %>New User<% end %></h3>

<div class="row">
  <div class="col-md-6">
  <form role="form" method="POST" action="/admin/users/edit<% if .update %>/id/<% .user.Id %><% end %>"> 
     
     <div class="form-group">   
        <label for="username">Username</label>      
        <input class="form-control" type="text" id="username" name="Username" value="<% if .user.Username %><% .user.Username %><% end %>">
      </div>

       <div class="form-group">
        <label for="full_name">Full Name</label>      
        <input class="form-control" type="text" id="full_name" name="FullName" value="<%if .user.FullName %><% .user.FullName %><% end %>"> 
      </div>

      <div class="form-group">
        <label for="email">Email</label>
        <input class="form-control" type="text" id="email" name="Email" value="<%if .user.Email %><% .user.Email %><% end %>">
      </div>

     <div class="form-group">
      <label for="role">Role</label>     
      <select name="Role" id="role" class="form-control" >
        <% $cur_role := .user.Role %>
        <% range $name,$val := .roles%>
          <option <% if equal $val $cur_role %>selected="selected" <% end %> value="<% $val %>"><% $name %></option>
        <% end %>
      </select>      
      </div>


     <div class="form-group">
      <label for="password">Password</label>      
      <input class="form-control" type="password" id="password" name="Password">  
      </div>

      <div class="form-group">
        <label class="control-label" for="password_again">Password (Again)</label>
        <input class="form-control" type="password" id="password_again" name="PasswordAgain"> 
      </div>

       <div class="form-group">
        <button type="submit" class="btn btn-primary">Save</button>
        <% if .update %><a class="btn btn-danger" href="/admin/users/delete/id/<% .user.Id %>">Delete</a><% end %>
        <a class="btn btn-default" href="/admin/users">Cancel</a>
      </div>

    </form>
  </div>



</div>
<% template "admin-footer.html" .%>