<script lang="ts">
	import TrashCan from 'svelte-material-icons/TrashCan.svelte';
	import Pencil from 'svelte-material-icons/Pencil.svelte';
	import Play from 'svelte-material-icons/Play.svelte';

	export let databaseUrl: string;
	export let databaseType: string;
	export let databaseIcon: string;

	export let editURL: string | undefined = undefined;
	export let deleteURL: string | undefined = undefined;
	export let viewURL: string | undefined = undefined;
</script>

<div
	class="card bg-base-100 text-lg font-bold flex flex-col gap-3 shadow-xl grow place-content-around mt-1 mb-1"
>
	<div class="card-body flex-col lg:flex-row">
		<div class="flex flex-row items-center gap-3 grow">
			<div class="flex flex-col items-center">
				<img src={databaseIcon} alt="Mysql" class="h-[30px] w-[30px] basis-full" />
				<div class="text-xs">{databaseType}</div>
			</div>
			<h2 class="overflow-auto break-all">{databaseUrl}</h2>
		</div>
		<div class="flex flex-col">
			{#if viewURL || editURL || deleteURL}
				<div class="lg:hidden divider divider-vertical" />
			{/if}
			<div class="flex flex-row justify-around align-top grow">
				{#if viewURL || editURL || deleteURL}
					<div class="hidden lg:flex divider divider-horizontal" />
				{/if}
				{#if viewURL}
					<a href={viewURL} type="button" class="mr-5 inline place-content-center text-blue-500">
						<Play size={25} />
					</a>
				{/if}
				{#if editURL}
					<a href={editURL} type="button" class="mr-5 inline place-content-center text-green-500">
						<Pencil size={25} />
					</a>
				{/if}
				{#if deleteURL}
					<a href={deleteURL} type="button" class="inline place-content-center text-red-500">
						<TrashCan size={25} />
					</a>
				{/if}
			</div>
		</div>
	</div>
</div>
