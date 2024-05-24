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
			<p class="py-6">
				Here you can edit a Git source. Git sources are used to scan repositories for secrets. You
				can configure the repository URL, and other settings here. All commits to the repository
				will be scanned for secrets.
			</p>
		</div>
		<div class="card w-108 bg-base-100 shadow-xl bordered">
			<div class="card-body">
				{#if form?.error}
					<div class="alert alert-error">{form.error}</div>
				{/if}
				<h2 class="card-title">Edit the Git Secret Scanner</h2>
				<form
					autocomplete="off"
					method="POST"
					use:enhance={() => {
						return async ({ update }) => {
							update({ reset: false });
						};
					}}
					class="flex flex-col gap-1"
				>
					<input type="hidden" name="projectName" value={data.project?.name} />
					<input type="hidden" name="organizationName" value={data.organization?.name} />
					<input type="hidden" name="sourceId" value={data.currentSource.id} />

					<BaseCreateDatabase
						fields={['privateKey', 'repository', 'username', 'password']}
						initialValues={{
							repository: data.currentSource.git_repository,
							privateKey: undefined,
							username: data.currentSource.username,
							password: data.currentSource.password
						}}
					/>
				</form>
			</div>
		</div>
	</div>
</div>
