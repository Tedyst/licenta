<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
</script>

<div class="flex items-center justify-center">
	<div class="card w-108 bg-base-100 shadow-xl bordered">
		<div class="card-body">
			<h2 class="card-title">Add a Source</h2>
			<p>Select the type of source you want to add:</p>
			<div class="p-2">
				<a
					href="/dashboard/{data.organization?.name}/{data.project?.name}/add-source/docker"
					class="btn btn-primary btn-block">Add a Docker source</a
				>
			</div>
			<div class="p-2">
				<a
					href="/dashboard/{data.organization?.name}/{data.project?.name}/add-source/git"
					class="btn btn-primary btn-block">Add a Git source</a
				>
			</div>
		</div>
	</div>
</div>
