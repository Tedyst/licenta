<script lang="ts">
	import { browser } from '$app/environment';
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import {
		JSONtoPublicKeyCredentialCreationOptions,
		PublicKeyCredentialToJSON
	} from '$lib/webauthn';
	import type { ActionData, PageData } from './$types';
	import TrashCan from 'svelte-material-icons/TrashCan.svelte';
	import aaguid from './aaguid.json';
	import WebauthnLogo from './webauthn_logo.png';

	export let data: PageData;
	export let form: ActionData;

	let name: string;
	let error: string | null = null;
	let formData: string;
	let f: HTMLFormElement;

	$: (async () => {
		if (!form?.setupCredential || !browser) {
			return;
		}
		const options = JSONtoPublicKeyCredentialCreationOptions(form.setupCredential);
		const credential = await navigator.credentials.create({ publicKey: options });
		console.log(credential);
		if (!credential) {
			error = 'Failed to create credential';
			return;
		}
		const credentialJSON = PublicKeyCredentialToJSON(credential);
		console.log(credentialJSON);
		const formData = new FormData(f);

		formData.set('data', JSON.stringify(credentialJSON));
		formData.set('name', name);

		await fetch(f.action, {
			method: f.method,
			body: formData
		});

		invalidateAll();
	})();

	let keys = data.user.webauthn_keys.map((key) => {
		let d = {
			info: { name: 'Generic Key', icon_dark: WebauthnLogo, icon_light: WebauthnLogo },
			...key
		};
		if (key.aaguid in aaguid) {
			let a = aaguid[key.aaguid as keyof typeof aaguid] as {
				name: string;
				icon_dark?: string;
				icon_light?: string;
			};
			d.info = {
				name: a.name,
				icon_dark: a?.icon_dark || '',
				icon_light: a?.icon_light || ''
			};
		}

		return d;
	});
</script>

<div class="container mx-auto p-4">
	{#if error}
		<div class="alert alert-error shadow-lg">
			<div>{error}</div>
		</div>
	{/if}

	<div class="form-control">
		<form action="?/start" use:enhance method="POST" class="grid grid-cols-1 gap-4">
			<label class="label" for="name">
				<span class="label-text">Add New Key</span>
			</label>
			<input
				type="text"
				name="name"
				placeholder="Name"
				required
				bind:value={name}
				class="input input-bordered"
			/>
			<button type="submit" class="btn btn-primary">Add key</button>
		</form>
	</div>

	<div class="divider">OR</div>

	<form action="?/finish" use:enhance method="POST" bind:this={f} class="hidden">
		<input type="hidden" name="name" bind:value={name} />
		<input type="hidden" name="data" value={formData} />
	</form>

	{#each keys || [] as key}
		<div
			class="card bg-base-100 text-lg font-bold flex flex-col gap-3 shadow-xl grow place-content-around mt-1 mb-1"
		>
			<div class="card-body flex-col lg:flex-row">
				<div class="flex flex-row items-center gap-3 grow">
					<div class="flex flex-col items-center">
						<img src={key.info.icon_light} alt="bitwarden" class="h-[30px] w-[30px] basis-full" />
						<div class="text-xs">{key.info.name}</div>
					</div>
					<h2 class="overflow-auto break-all">{key.name}</h2>
				</div>
				<div class="flex flex-col">
					<div class="lg:hidden divider divider-vertical" />
					<div class="flex flex-row justify-around align-top grow">
						<div class="hidden lg:flex divider divider-horizontal" />
						<a href={'asd'} type="button" class="inline place-content-center text-red-500">
							<TrashCan size={25} />
						</a>
					</div>
				</div>
			</div>
		</div>
	{/each}
</div>
