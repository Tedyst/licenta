<script lang="ts">
	import type { components } from '$lib/api/v1';
	import TrashCan from 'svelte-material-icons/TrashCan.svelte';

	export let organization: components['schemas']['Organization'];
	export let project: components['schemas']['Project'];
</script>

<a href="/dashboard/{organization.name}/{project.name}">
	<div class="card ml-4 h-30 max-w-full mt-2 bg-base-100">
		<div class="card-body flex-row pr-4">
			<div class="flex flex-grow align-middle items-start flex-col self-center">
				Project
				<div class="ml-4 text-2xl text-info">{project.name}</div>
			</div>
			<div class="mr-0 inline-grid grid-cols-2 grid-rows-1 grow-0 gap-2">
				<div class="inline-grid w-max gap-4">
					<div class="stat-title text-sm">Scans</div>
					<div class="stat-value text-success text-base">0</div>
				</div>
			</div>
			<div class="divider divider-horizontal" />
			<a
				href="/dashboard/{organization.name}/{project.name}/delete"
				type="button"
				class="mr-5 inline place-content-center text-red-500"
			>
				<TrashCan size={25} />
			</a>
		</div>
	</div>
</a>
