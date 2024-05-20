<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { enhance } from '$app/forms';
	import BaseCreateDatabase from '$lib/dashboard/BaseCreateDatabase.svelte';
	export let data: PageData;
	export let form: ActionData;
</script>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left">
			<p class="py-6">Here you can add a Docker image</p>
		</div>
		<div class="card w-108 bg-base-100 shadow-xl bordered">
			<div class="card-body">
				{#if form?.error}
					<div class="alert alert-error">{form.error}</div>
				{/if}
				<h2 class="card-title">Add a Docker Image</h2>
				<form autocomplete="off" method="POST" use:enhance class="flex flex-col gap-1">
					<input type="hidden" name="projectId" value={data.project?.id} />
					<input type="hidden" name="projectName" value={data.project?.name} />
					<input type="hidden" name="organizationName" value={data.organization?.name} />
					<BaseCreateDatabase fields={['image', 'username', 'password']} />
				</form>
			</div>
		</div>
	</div>
</div>
