<% template "admin-header.html" .%>

<span class="pull-right"><a href="/admin/users">Back to list &raquo;</a></span>
<h1>Manage Users</h1>

<form class="form-horizontal" method="POST" action="/admin/users/edit<% if .update %>/id/<% .user.Id %><% end %>">
  <fieldset>
   <legend><% if .update %>Edit User<% else %>New User<% end %></legend>
    <div class="control-group">
      <label class="control-label" for="username">Username</label>
      <div class="controls">
        <input type="text" class="span4" id="username" name="Username" value="<% if .user.Username %><% .user.Username %><% end %>">        
      </div>
    </div>

    <div class="control-group">
      <label class="control-label" for="full_name">Full Name</label>
      <div class="controls">
        <input type="text" class="span4" id="full_name" name="FullName" value="<%if .user.FullName %><% .user.FullName %><% end %>">        
      </div>
    </div>

    <div class="control-group">
      <label class="control-label" for="email">Email</label>
      <div class="controls">
        <input type="text" class="span4" id="email" name="Email" value="<%if .user.Email %><% .user.Email %><% end %>">        
      </div>
    </div>

    <div class="control-group">
      <label class="control-label" for="role">Role</label>
      <div class="controls">
          <select class="span4" name="Role" id="role">
            <% $cur_role := .user.Role %>
            <% range $name,$val := .roles%>
              <option <% if eq $val $cur_role %>selected="selected" <% end %> value="<% $val %>"><% $name %></option>
            <% end %>
          </select>
      </div>
    </div>


    <div class="control-group">
      <label class="control-label" for="password">Password</label>
      <div class="controls">
        <input type="password" class="span4" id="password" name="Password">
        
      </div>
    </div>

    <div class="control-group">
      <label class="control-label" for="password_again">Password (Again)</label>
      <div class="controls">
          <input type="password" class="span4" id="password_again" name="PasswordAgain"> 
      </div>
    </div>

    <div class="form-actions">
      <button type="submit" class="btn btn-primary">Save</button>
      <% if .update %><a class="btn btn-danger" href="/admin/users/delete/id/<% .user.Id %>">Delete</a><% end %>
      <a class="btn" href="/admin/users">Cancel</a>
    </div>

  </fieldset>
</form>

<% template "admin-footer.html" .%>