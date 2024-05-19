<script context="module" lang="ts">
	export type Field = 'hostname' | 'port' | 'username' | 'password' | 'database';
</script>

<script lang="ts">
	import Key from 'svelte-material-icons/Key.svelte';
	import Account from 'svelte-material-icons/Account.svelte';
	import Server from 'svelte-material-icons/Server.svelte';
	import Podcast from 'svelte-material-icons/Podcast.svelte';
	import Database from 'svelte-material-icons/Database.svelte';

	export let fields: Field[] = ['hostname', 'port', 'username', 'password', 'database'];
	export let initialValues: Record<Field, string | undefined> = {
		hostname: '',
		port: '',
		username: '',
		password: '',
		database: ''
	};
</script>

{#if fields.includes('hostname')}
	<label class="input input-bordered flex items-center gap-2">
		<Server />
		<input
			type="text"
			class="grow bg-base-100"
			placeholder="Hostname"
			autocomplete="off"
			name="hostname"
			value={initialValues.hostname || ''}
		/>
	</label>
{/if}
{#if fields.includes('port')}
	<label class="input input-bordered flex items-center gap-2">
		<Podcast />
		<input
			type="number"
			class="grow bg-base-100"
			placeholder="Port"
			autocomplete="off"
			name="port"
			value={initialValues.port || ''}
		/>
	</label>
{/if}
{#if fields.includes('username')}
	<label class="input input-bordered flex items-center gap-2">
		<Account />
		<input
			type="text"
			class="grow bg-base-100"
			placeholder="Username"
			autocomplete="off"
			name="username"
			value={initialValues.username || ''}
		/>
	</label>
{/if}
{#if fields.includes('password')}
	<label class="input input-bordered flex items-center gap-2">
		<Key />
		<input
			type="password"
			class="grow bg-base-100"
			placeholder="Password"
			autocomplete="off"
			name="password"
			value={initialValues.password || 'pass'}
		/>
	</label>
{/if}
{#if fields.includes('database')}
	<label class="input input-bordered flex items-center gap-2">
		<Database />
		<input
			type="text"
			class="grow bg-base-100"
			placeholder="Database Name"
			autocomplete="off"
			name="database"
			value={initialValues.database || ''}
		/>
	</label>
{/if}

<button class="btn btn-primary">Create</button>
