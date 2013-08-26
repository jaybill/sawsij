<% template "header.html" . %>

 <div class="row">
  <div class="col-md-4">
  <form class="" method="post" action="/login" role="form">  
    
    <div class="form-group">
      <label for="username" class="control-label">Username</label>            
      <input type="text" class="form-control" name="username" id="username" <%if .username %>value="<% .username %>"<% end %> >              
    </div>

    <div class="form-group">
      <label for="password" class="control-label">Password</label>          
      <input type="password" class="form-control" name="password" id="password">                  
    </div>

  <% if .dest %><input type="hidden" id="dest" name="dest" value="<% .dest %>"/><% end %> 
  
  <div class="form-group">    
    <button type="submit" class="btn btn-primary">Log In</button>    
  </div>

  </div>
  <div class="col-md-8">
  </div>
</form> 

<% template "footer.html" . %>