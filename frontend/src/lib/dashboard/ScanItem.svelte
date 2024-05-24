<script lang="ts">
	import type { components } from '$lib/api/v1';
	import Play from 'svelte-material-icons/Play.svelte';

	export let scan: components['schemas']['Scan'];

	export let scanIcon: string;
	export let scanName: string;

	export let viewURL: string | undefined = undefined;

	const scanColor =
		scan.ended_at === '0001-01-01T00:00:00Z'
			? 'bg-base-200'
			: scan.maximum_severity === 3
				? 'bg-red-200'
				: scan.maximum_severity === 2
					? 'bg-yellow-200'
					: scan.maximum_severity === 1
						? 'bg-blue-200'
						: 'bg-base-200';

	const severityName =
		scan.maximum_severity === 4
			? 'Critical'
			: scan.maximum_severity === 3
				? 'High'
				: scan.maximum_severity === 2
					? 'Medium'
					: scan.maximum_severity === 1
						? 'Low'
						: 'Info';
</script>

<div
	class="card {scanColor} text-lg font-bold flex flex-col gap-3 shadow-xl grow place-content-around mt-1 mb-1"
>
	<div class="card-body flex-col lg:flex-row">
		<div class="flex flex-row items-center gap-3 grow justify-between">
			<div class="flex flex-row items-center gap-3">
				<div class="flex flex-col items-center">
					<img src={scanIcon} alt="Mysql" class="h-[30px] w-[30px] basis-full" />
					<div class="text-xs">{scanName}</div>
				</div>
				<div class="flex flex-col grow">
					<h2 class="overflow-auto break-all">{scanName} Scan ID {scan.id}</h2>
					<h2 class="overflow-auto break-all text-sm">Severity {severityName}</h2>
				</div>
			</div>
			<div class="flex flex-row items-center">
				<div class="flex flex-col items-center justify-stretch">
					<div class="text-sm flex justify-between grow">Started {scan.created_at}</div>
					{#if scan.ended_at !== '0001-01-01T00:00:00Z'}
						<div class="text-sm flex justify-between grow">Finished {scan.ended_at}</div>
					{:else}
						<div class="text-sm flex justify-between grow">Running</div>
					{/if}
				</div>
				{#if viewURL}
					<div class="lg:hidden divider divider-vertical" />
					<div class="flex flex-row justify-around align-top grow">
						<div class="hidden lg:flex divider divider-horizontal" />
						<a href={viewURL} type="button" class="mr-5 inline place-content-center text-blue-500">
							<Play size={25} />
						</a>
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>
