<script lang="ts">
	import type { components } from '$lib/api/v1';

	import TrashCan from 'svelte-material-icons/TrashCan.svelte';
	import Pencil from 'svelte-material-icons/Pencil.svelte';

	import MysqlIcon from '$lib/images/mysql-icon.svg';
	import EditMysqlDatabase from './EditMysqlDatabase.svelte';
	import client from '$lib/client';
	import { currentMysqlDatabases } from '$lib/stores';
	import { toast } from 'svelte-daisy-toast';

	export let mysqlDatabase: components['schemas']['MysqlDatabase'];

	let editComponent: HTMLDivElement;
	let modal: HTMLDialogElement;

	const toggleEditComponent = () => {
		editComponent.classList.toggle('hidden');
		editComponent.classList.toggle('flex');
	};

	const deleteDatabase = (id: number) => {
		client.DELETE('/mysql/{id}', { params: { path: { id } } }).then((response) => {
			if (response.error) {
				console.error(response.error);
				return;
			}
			$currentMysqlDatabases = $currentMysqlDatabases.filter((db) => db.id !== id);
			modal.close();
			toast({
				closable: true,
				duration: 5000,
				message: 'Database deleted successfully',
				type: 'success'
			});
		});
	};
</script>

<div
	class="card bg-base-100 text-lg font-bold flex flex-col gap-3 shadow-xl grow place-content-around mt-1 mb-1"
>
	<div class="card-body flex-col lg:flex-row">
		<div class="flex flex-row items-center gap-3 grow">
			<div class="flex flex-col">
				<img src={MysqlIcon} alt="Mysql" class="h-[30px] w-[30px]" />
				<div class="text-xs">MySQL</div>
			</div>
			<h2 class="overflow-auto">
				mysql://{mysqlDatabase.username}@****:{mysqlDatabase.host}:{mysqlDatabase.port}/{mysqlDatabase.database_name}
			</h2>
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
		<EditMysqlDatabase {mysqlDatabase} />
	</div>
</div>

<dialog class="modal" bind:this={modal}>
	<div class="modal-box">
		<h3 class="font-bold text-lg">
			Are you sure that you want to delete the database <b
				>mysql://{mysqlDatabase.username}@****:{mysqlDatabase.host}:{mysqlDatabase.port}/{mysqlDatabase.database_name}</b
			>?
		</h3>
		<p class="py-4">Press ESC key or click the button below to close</p>
		<div class="modal-action">
			<form method="dialog">
				<button class="btn">No</button>
				<button
					class="btn bg-red-500 text-black"
					on:click|stopPropagation|preventDefault={() => deleteDatabase(mysqlDatabase.id)}
					>Yes</button
				>
			</form>
		</div>
	</div>
</dialog>
