<script lang="ts">
	import type { components } from '$lib/api/v1';
	import client from '$lib/client';
	import { currentProject, currentMysqlDatabases } from '$lib/stores';

	import Key from 'svelte-material-icons/Key.svelte';
	import Account from 'svelte-material-icons/Account.svelte';
	import Server from 'svelte-material-icons/Server.svelte';
	import Podcast from 'svelte-material-icons/Podcast.svelte';
	import Database from 'svelte-material-icons/Database.svelte';
	import { toast } from 'svelte-daisy-toast';

	let onSubmit = (e: SubmitEvent) => {
		if (!$currentProject) return;

		const formData = new FormData(e.target as HTMLFormElement);

		const hostname = formData.get('hostname') as string;
		const port = formData.get('port') as string;
		const username = formData.get('username') as string;
		const password = formData.get('password') as string;
		const database = formData.get('database') as string;

		client
			.POST('/mysql', {
				body: {
					database_name: database,
					host: hostname,
					password,
					port: parseInt(port),
					project_id: $currentProject?.id,
					username
				}
			})
			.then((response) => {
				if (!response.data?.mysql_database) return;
				$currentMysqlDatabases = [...$currentMysqlDatabases, response.data?.mysql_database];
				toast({
					closable: true,
					duration: 5000,
					message: 'Database created successfully',
					type: 'success'
				});
			});
	};
</script>

<form autocomplete="off" on:submit|preventDefault={onSubmit}>
	<div class="card w-96 bg-base-100 shadow-xl">
		<div class="card-body">
			<label class="input input-bordered flex items-center gap-2">
				<Server />
				<input
					type="text"
					class="grow bg-base-100"
					placeholder="Hostname"
					autocomplete="off"
					name="hostname"
				/>
			</label>
			<label class="input input-bordered flex items-center gap-2">
				<Podcast />
				<input
					type="number"
					class="grow bg-base-100"
					placeholder="Port"
					autocomplete="off"
					name="port"
				/>
			</label>
			<label class="input input-bordered flex items-center gap-2">
				<Account />
				<input
					type="text"
					class="grow bg-base-100"
					placeholder="Username"
					autocomplete="off"
					name="username"
				/>
			</label>
			<label class="input input-bordered flex items-center gap-2">
				<Key />
				<input
					type="password"
					class="grow bg-base-100"
					placeholder="Password"
					autocomplete="off"
					name="password"
				/>
			</label>
			<label class="input input-bordered flex items-center gap-2">
				<Database />
				<input
					type="text"
					class="grow bg-base-100"
					placeholder="Database Name"
					autocomplete="off"
					name="database"
				/>
			</label>

			<button class="btn btn-primary">Create</button>
		</div>
	</div>
</form>
