<% template "header.html" . %>

<h2>Login</h2>

<%if .failed %><div class="alert alert-error">Login Failed</div><% end %>

<form class="" method="post" action="/login">
  
    
    <div class="control-group" >
      <label class="control-label" for="username">Username</label>
      <div class="controls">
        <input type="text" class="input-xlarge" name="username" id="username" <%if .username %>value="<% .username %>"<% end %> >        
      </div>
    </div>

    <div class="control-group">
      <label class="control-label" for="password">Password</label>
      <div class="controls">
        <input type="password" class="input-xlarge" name="password" id="password">        
      </div>
    </div>
  <% if .dest %><input type="hidden" id="dest" name="dest" value="<% .dest %>"/><% end %> 
    <button type="submit" class="btn btn-primary">Log In</button>
  
</form>

<% template "footer.html" . %>