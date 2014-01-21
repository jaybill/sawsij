<% template "header.html" .%>

<div class="jumbotron">
 <h1>Welcome!</h1>
  <p>Your new sawsij application is up and running.</p>
</div>

<div class="row">

	<div class="span6">
		<h2>Key Files</h2>
		<p>Here's a list of some key files and directories in your application.</p>

		<ul>
			<li><b>src/{{.name}}server/{{ .name }}server.go</b><br />
				The main application server source. This is where the <b>main()</b> function is.
				Generally, this is where you'll add routes and handlers.
			</li>
			<li><b>etc/config.yaml</b><br />
				The primary configuration file. Controls things like what port your app answers on
				and your database parameters.
			</li>
			<li><b>templates/</b><br />
				The html templates for your application. The template files are named according to the URL pattern for the route.
			</li>
			<li><b>static/</b><br />
				Where static content lives. Things like images, CSS files and Javascript.
			</li>
			<li><b>templates/index.html</b><br />
				The html template for the page you're currently viewing. You can delete the contents and replace it with your own.
			</li>			
		</ul>
	</div>
	<div class="span6">		
		<h2>Documentation</h2>
		<p>Here's all the relevant documentation.</p>
		<li><a href="https://bitbucket.org/jaybill/sawsij/wiki/Home">Documentation Wiki</a></li>
		<li><a href="http://go.pkgdoc.org/bitbucket.org/jaybill/sawsij/framework">API Documentation</a></li>
		<li><a href="http://golang.org/ref/">Go Documentation</a></li>
		<li><a href="http://golang.org/pkg/text/template/">Template Documentation</a></li>
		<li><a href="http://getbootstrap.com/">Bootstrap</a></li>
	</div>	
</div>


<% template "footer.html" .%>