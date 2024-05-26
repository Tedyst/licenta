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
				Here you can edit a PostgreSQL database from project. Make sure that the PostgreSQL server
				is running and that you have the necessary credentials to connect to it.
				<br />
				In case you have a REMOTE project, make sure you use internal IP addresses to connect to the
				database (the address must be accessible from the server running the worker).
				<br />
				Also, make sure that you have the necessary permissions to see the pg_users table using the credentials
				entered below. We recommend using a dedicated user for this purpose, with no other permissions.
			</p>
		</div>
		<div class="card w-108 bg-base-100 shadow-xl bordered">
			<div class="card-body">
				{#if form?.error}
					<div class="alert alert-error">{form.error}</div>
				{/if}
				<h2 class="card-title">Edit the PostgreSQL database</h2>
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
					<input type="hidden" name="databaseId" value={data.currentDatabase?.id} />

					<BaseCreateDatabase
						fields={['hostname', 'port', 'username', 'password', 'database']}
						initialValues={{
							hostname: data.currentDatabase?.host,
							port: data.currentDatabase?.port.toString(),
							username: data.currentDatabase?.username,
							password: data.currentDatabase?.password,
							database: data.currentDatabase?.database_name
						}}
						buttonLabel="Edit"
					/>
				</form>
			</div>
		</div>
	</div>
</div>
