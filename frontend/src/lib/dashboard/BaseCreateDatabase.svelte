<script context="module" lang="ts">
	export type Field =
		| 'hostname'
		| 'port'
		| 'username'
		| 'password'
		| 'database'
		| 'repository'
		| 'privateKey'
		| 'image';
</script>

<script lang="ts">
	import Key from 'svelte-material-icons/Key.svelte';
	import Account from 'svelte-material-icons/Account.svelte';
	import Server from 'svelte-material-icons/Server.svelte';
	import Podcast from 'svelte-material-icons/Podcast.svelte';
	import Database from 'svelte-material-icons/Database.svelte';
	import LinkVariant from 'svelte-material-icons/LinkVariant.svelte';

	export let fields: Field[] = ['hostname', 'port', 'username', 'password', 'database'];
	export let initialValues: Partial<Record<Field, string | undefined>> = {};
</script>

{#if fields.includes('repository')}
	<label class="input input-bordered flex items-center gap-2">
		<LinkVariant />
		<input
			type="text"
			class="grow bg-base-100"
			placeholder="Repository"
			autocomplete="off"
			name="repository"
			value={initialValues?.repository || ''}
		/>
	</label>
{/if}
{#if fields.includes('image')}
	<label class="input input-bordered flex items-center gap-2">
		<LinkVariant />
		<input
			type="text"
			class="grow bg-base-100"
			placeholder="Image"
			autocomplete="off"
			name="image"
			value={initialValues?.image || ''}
		/>
	</label>
{/if}
{#if fields.includes('hostname')}
	<label class="input input-bordered flex items-center gap-2">
		<Server />
		<input
			type="text"
			class="grow bg-base-100"
			placeholder="Hostname"
			autocomplete="off"
			name="hostname"
			value={initialValues?.hostname || ''}
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
			value={initialValues?.port || ''}
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
			value={initialValues?.username || ''}
		/>
	</label>
{/if}
{#if fields.includes('password')}
	<label class="input input-bordered flex items-center gap-2">
		<Key />
		<input
			type="password"
			class="grow bg-base-100"
			placeholder="•••••••••••••"
			autocomplete="off"
			name="password"
			value={initialValues?.password || ''}
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
			value={initialValues?.database || ''}
		/>
	</label>
{/if}
{#if fields.includes('privateKey')}
	<label class="textarea textarea-bordered flex items-center gap-2">
		<Key />
		<textarea
			class="textarea grow bg-base-100 min-h-10"
			placeholder="Private Key"
			autocomplete="off"
			name="privateKey"
			value={initialValues?.privateKey || ''}
		/>
	</label>
{/if}

<button class="btn btn-primary">Create</button>
