<script lang="ts">
	import { browser } from '$app/environment';
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import {
		JSONtoPublicKeyCredentialCreationOptions,
		PublicKeyCredentialToJSON
	} from '$lib/webauthn';
	import type { ActionData, PageData } from './$types';

	export let data: PageData;
	export let form: ActionData;

	let name: string;
	let error: string | null = null;
	let formData: string;
	let f: HTMLFormElement;

	$: (async () => {
		console.log(form);
		if (!form?.setupCredential || !browser) {
			return;
		}
		const options = JSONtoPublicKeyCredentialCreationOptions(form.setupCredential);
		const credential = await navigator.credentials.create({ publicKey: options });
		if (!credential) {
			error = 'Failed to create credential';
			return;
		}
		const credentialJSON = PublicKeyCredentialToJSON(credential);
		const formData = new FormData(f);

		formData.set('data', JSON.stringify(credentialJSON));
		formData.set('name', name);

		const response = await fetch(f.action, {
			method: f.method,
			body: formData
		});

		invalidateAll();
	})();
</script>

{#if error}
	<p>{error}</p>
{/if}

{#if form?.setupCredential}
	{form?.setupCredential}
{/if}

<form action="?/start" use:enhance method="POST">
	<input type="text" name="name" placeholder="Name" required bind:value={name} />
	<button type="submit" class="btn btn-primary">Add key</button>
</form>

<form action="?/finish" use:enhance method="POST" bind:this={f}>
	<input type="hidden" name="name" bind:value={name} />
	<input type="hidden" name="data" value={formData} />
</form>

Keys:
{#each data?.user?.webauthn_keys || [] as key}
	{key.id} - {key.name}
{/each}
