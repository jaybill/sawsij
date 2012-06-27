<% template "header.html" %>

<h2>Login</h2>

<%if .failed %><div class="alert alert-error">Login Failed</div><% end %>
<form class="form" method="post" action="/login">
 
<label for="username">Username
<input type="text" id="username" name="username" <%if .username %>value="<% .username %>"/><% end %></label>

<label for="password">Password
<input type="password" id="password" name="password"/></label>

<% if .dest %><input type="hidden" id="dest" name="dest" value="<% .dest %>"/><% end %>

<button type="submit" class="btn btn-primary">Log In</button>
 
</form>

<% template "footer.html" %>