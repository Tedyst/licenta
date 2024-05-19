<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { enhance } from '$app/forms';
	export let data: PageData;
	export let form: ActionData;
</script>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left">
			<p class="py-6">Here you can delete a PostgreSQL database from project.</p>
		</div>
		<div class="card w-108 bg-base-100 shadow-xl bordered">
			<div class="card-body">
				{#if form?.error}
					<div class="alert alert-error">{form.error}</div>
				{/if}
				<h2 class="card-title">Delete the PostgreSQL database</h2>
				<p class="text-center">
					Are you sure you want to delete the PostgreSQL database? This action is irreversible.
				</p>
				<form autocomplete="off" method="POST" use:enhance class="flex flex-col gap-1">
					<input type="hidden" name="databaseId" value={data.currentDatabase?.id} />
					<input type="hidden" name="projectName" value={data.project?.name} />
					<input type="hidden" name="organizationName" value={data.organization?.name} />
					<button type="submit" class="btn btn-error">Delete</button>
				</form>
			</div>
		</div>
	</div>
</div>
