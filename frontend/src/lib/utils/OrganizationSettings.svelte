<script lang="ts">
	import type { components } from '$lib/api/v1';
	import client from '$lib/client';
	import { toast } from 'svelte-daisy-toast';

	export let organization: components['schemas']['Organization'];

	let modal: HTMLDialogElement;

	let deleteOrganization = () => {
		client
			.DELETE('/organizations/{id}', { params: { path: { id: organization.id } } })
			.then((response) => {
				if (response.data?.success) {
					window.location.href = '/dashboard';
					toast({
						closable: true,
						duration: 5000,
						message: 'Organization deleted successfully',
						title: 'Success',
						type: 'success'
					});
				} else {
					toast({
						closable: true,
						duration: 5000,
						message: response.error?.message || 'Could not delete organization',
						title: 'Error',
						type: 'error'
					});
				}
			});
	};
</script>

<div class="grid flex-grow place-content-center h-auto flex-1">
	<h1 class="text-3xl font-bold">Organization Settings</h1>
	<form class="card-body" on:submit|preventDefault={() => modal.showModal()}>
		<div class="form-control mt-6">
			<button class="btn btn-error">Delete organization</button>
		</div>
	</form>
</div>

<dialog class="modal" bind:this={modal}>
	<div class="modal-box">
		<h3 class="font-bold text-lg">
			Are you sure that you want to delete the organization <b>{organization.name}</b>?
		</h3>
		<p class="py-4">Press ESC key or click the button below to close</p>
		<div class="modal-action">
			<form method="dialog">
				<button class="btn">No</button>
				<button class="btn bg-red-500 text-black" on:click={deleteOrganization}>Yes</button>
			</form>
		</div>
	</div>
</dialog>
