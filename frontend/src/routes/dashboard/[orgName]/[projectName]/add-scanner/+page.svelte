<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
</script>

<div class="flex items-center justify-center">
	<div class="card w-108 bg-base-100 shadow-xl bordered">
		<div class="card-body">
			<h2 class="card-title">Add a Database</h2>
			<p>Select the type of database you want to add:</p>
			<div class="p-2">
				<a
					href="/dashboard/{data.organization?.name}/{data.project?.name}/add-scanner/postgresql"
					class="btn btn-primary btn-block">Add a PostgreSQL database</a
				>
			</div>
			<div class="p-2">
				<a
					href="/dashboard/{data.organization?.name}/{data.project?.name}/add-scanner/mysql"
					class="btn btn-primary btn-block">Add a MySQL database</a
				>
			</div>
		</div>
	</div>
</div>
