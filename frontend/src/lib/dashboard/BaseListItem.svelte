<script lang="ts">
	import TrashCan from 'svelte-material-icons/TrashCan.svelte';
	import Pencil from 'svelte-material-icons/Pencil.svelte';

	export let databaseUrl: string;
	export let databaseType: string;
	export let databaseIcon: string;
	export let databaseID: number;
	export let deleteAction: (id: number) => void;

	let editComponent: HTMLDivElement;
	let modal: HTMLDialogElement;

	const toggleEditComponent = () => {
		editComponent.classList.toggle('hidden');
		editComponent.classList.toggle('flex');
	};

	const deleteDatabase = () => {
		deleteAction(databaseID);
		modal.close();
	};
</script>

<div
	class="card bg-base-100 text-lg font-bold flex flex-col gap-3 shadow-xl grow place-content-around mt-1 mb-1"
>
	<div class="card-body flex-col lg:flex-row">
		<div class="flex flex-row items-center gap-3 grow">
			<div class="flex flex-col items-center">
				<img src={databaseIcon} alt="Mysql" class="h-[30px] w-[30px]" />
				<div class="text-xs">{databaseType}</div>
			</div>
			<h2 class="overflow-auto">{databaseUrl}</h2>
		</div>
		<div class="flex flex-col">
			<div class="lg:hidden divider divider-vertical" />
			<div class="flex flex-row justify-around align-top grow">
				<div class="hidden lg:flex divider divider-horizontal" />
				<button
					type="button"
					class="mr-5 inline place-content-center text-green-500"
					on:click|preventDefault={toggleEditComponent}
				>
					<Pencil size={25} />
				</button>
				<button
					type="button"
					class="inline place-content-center text-red-500"
					on:click|preventDefault={() => modal.showModal()}
				>
					<TrashCan size={25} />
				</button>
			</div>
		</div>
	</div>
	<div class="hidden justify-center align-middle pb-8" bind:this={editComponent}>
		<slot name="editbox" />
	</div>
</div>

<dialog class="modal" bind:this={modal}>
	<div class="modal-box">
		<h3 class="font-bold text-lg">
			Are you sure that you want to delete the database <b>{databaseUrl}</b>?
		</h3>
		<p class="py-4">Press ESC key or click the button below to close</p>
		<div class="modal-action">
			<form method="dialog">
				<button class="btn">No</button>
				<button
					class="btn bg-red-500 text-black"
					on:click|stopPropagation|preventDefault={() => deleteDatabase()}>Yes</button
				>
			</form>
		</div>
	</div>
</dialog>
