package menus

import (
	"net/http"
	"path"
	"time"

	"github.com/bradleypeabody/gocaveman"
)

var ErrNotFound = gocaveman.ErrNotFound

var modTime = time.Now()

func NewDefaultAdminMenusViewFS() http.FileSystem {
	return NewAdminMenusViewFS("/admin/menus")
}

func NewAdminMenusViewFS(prefix string) http.FileSystem {
	return gocaveman.NewHTTPFuncFS(func(name string) (http.File, error) {
		name = path.Clean("/" + name)
		if name == prefix+"/index.gohtml" {
			return gocaveman.NewHTTPBytesFile(name, modTime, adminMenusListViewBytes), nil
		}
		if name == prefix+"/edit.gohtml" {
			return gocaveman.NewHTTPBytesFile(name, modTime, adminMenusEditViewBytes), nil
		}
		return nil, ErrNotFound
	})
}

var adminMenusListViewBytes = []byte(`{{template "/admin-page.gohtml" .}}
{{define "body"}}

{{$req := .Value "req"}}

<script src="https://unpkg.com/vue"></script>

<h1>List View</h1>

<div id="main_error" class="alert alert-danger collapse" role="alert"></div>

<ul id="list_view" class="collapse">
	<li v-for="menu in menus"><a :href="menu._path">${menu.id}</a></li>
</ul>

<div>
	<a href="{{$req.URL.Path}}/edit" class="btn btn-primary">Add Menu</a>
</div>

<script>

document.addEventListener("DOMContentLoaded", function(event) { 

	fetch('/api/menus').then(function(res) {

		if (res.status == 200) {
			return res.json();
		} else {
			console.log("Error fetching menu data:");
			console.log(res);
			var e = document.querySelector('#main_error');
			e.innerHTML = res.statusText;
			e.classList.remove('collapse');
		}

	}).then(function(data) {

		var prefix = '{{$req.URL.Path}}/edit?id=';

		for (var i = 0; i < data.length; i++) {
			data[i]._path = prefix + data[i].id;
		}

		var list_view = new Vue({
			delimiters: ['${', '}'],
			el: '#list_view',
			data: {
				menus: data,
			},
		});

		document.querySelector('#list_view').classList.remove('collapse');

	});

});

</script>

{{end}}
`)

var adminMenusEditViewBytes = []byte(`{{template "/admin-page.gohtml" .}}
{{define "body"}}

{{$req := .Value "req"}}
{{$id := $req.URL.Query.Get "id"}}

<h1>Edit Menu {{if $id}}"{{$id}}"{{end}}</h1>
<a href="./">Back to Menu list</a>

<div id="main_error" class="alert alert-danger collapse" role="alert"></div>

<form id="main_form" class="collapse">

	<div class="form-group">
		<label for="menu_id">Menu ID</label>
		<input type="text" class="form-control" id="menu_id" aria-describedby="identifier" placeholder="Enter ID">
		<small id="idHelp" class="form-text text-muted">This uniquely identifies the menu.</small>
	</div>

	<div class="form-group">
		<label for="menu_json">Menu JSON</label>
		<textarea class="form-control" id="menu_json" rows="8"></textarea>
	</div>

	<div>
		<button id="save_btn" type="button" class="btn btn-primary">Save</button>
	</div>

</form>

<script>

document.addEventListener("DOMContentLoaded", function(event) {

	var id = {{$id}};

	// if existing one then load it
	if (id && id.length) {

		fetch('/api/menus/'+id).then(function(res) {
			
			if (res.status == 200) {
				return res.json();
			} else {
				console.log("Error fetching menu data:");
				console.log(res);
				var e = document.querySelector('#main_error');
				e.innerHTML = res.statusText;
				e.classList.remove('collapse');
			}

		}).then(function(data) {

			console.log(data);

			document.querySelector('#menu_id').value = id;
			document.querySelector('#menu_id').disabled = true;
			if (data) {
				document.querySelector('#menu_json').value = JSON.stringify(data, null, 4);
			}

			document.querySelector('#main_form').classList.remove('collapse');

		});

	} else {

		document.querySelector('#main_form').classList.remove('collapse');

	}


	document.querySelector('#save_btn').addEventListener('click', function(e) {

		var e = document.querySelector('#main_error');

		var id = document.querySelector('#menu_id').value;
		if (!id) {
			e.innerHTML = "ID is required!";
			e.classList.remove('collapse');
			return;
		}

		var data = {};
		try {
			data = JSON.parse(document.querySelector('#menu_json').value);
		} catch(err) {
			console.log(err);
			e.innerHTML = "JSON Parse Error: "+err;
			e.classList.remove('collapse');
			return;
		}

		e.classList.add('collapse');

		fetch('/api/menus/'+id, {
			method: 'post',
			body: JSON.stringify(data),
		}).then(function(res) {
			if (res.status != 200) {
				console.log(res);
				e.innerHTML = "Save Error: "+res.statusText;
				e.classList.remove('collapse');
				return;
			} else {
				alert("Saved!");
			}
		});

	});

});

</script>

{{end}}
`)
