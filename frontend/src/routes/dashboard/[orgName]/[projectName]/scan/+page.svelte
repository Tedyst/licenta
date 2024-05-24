<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';
	import { invalidate } from '$app/navigation';
	export let data: PageData;

	const getResultColor = (severity: number) => {
		if (severity === 3) {
			return 'bg-red-200';
		} else if (severity === 2) {
			return 'bg-yellow-200';
		} else if (severity === 1) {
			return 'bg-blue-200';
		} else {
			return 'bg-base-200';
		}
	};

	const getNameSeverity = (severity: number) => {
		if (severity === 3) {
			return 'High';
		} else if (severity === 2) {
			return 'Medium';
		} else if (severity === 1) {
			return 'Warning';
		} else {
			return 'Info';
		}
	};

	onMount(() => {
		const interval = setInterval(() => {
			if (data.scan?.scan?.status === 2) return;
			invalidate('app:currentScan');
		}, 1000);

		return () => {
			clearInterval(interval);
		};
	});
</script>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left justify-stretch grow">
			{#if data.scan?.scan?.status === 0}
				<p class="py-6 max-w-md">
					Scan not started yet. Waiting for a worker to start processing it. The task might also be
					dead, in which case it might be needed to start another one
				</p>
			{:else if data.scan?.scan?.status === 1}
				<p class="py-6">Scan in progress</p>
			{:else if data.scan?.scan?.status === 2}
				<p class="py-6">Scan finished</p>
			{:else if data.scan?.scan?.status === 3}
				<p class="py-6">Scan is queued to a Project specific worker</p>
			{:else if data.scan?.scan?.status === 4}
				<p class="py-6">
					Scan is checking the public access before scheduling the task to a project specific worker
				</p>
			{/if}
		</div>
	</div>
</div>

<div class="flex flex-col justify-stretch grow">
	<div class="divider">Bruteforce Results</div>
	{#each data?.scan?.bruteforceResults as bruteforceResult}
		<div
			class="card {bruteforceResult.password
				? 'bg-red-200'
				: 'bg-green-200'} text-lg font-bold flex flex-col gap-3 shadow-xl grow place-content-around mt-1 mb-1"
		>
			<div class="card-body flex-col lg:flex-row">
				<div class="flex flex-row items-center gap-3 grow justify-between">
					<div class="flex flex-row items-center gap-3">
						<div class="flex flex-col grow">
							<h2 class="overflow-auto break-all">User {bruteforceResult.username}</h2>
							<h2 class="overflow-auto break-all text-sm">
								Tried {bruteforceResult.tried} passwords out of {bruteforceResult.total}
							</h2>
						</div>
					</div>
					{#if bruteforceResult.password}
						<div class="flex flex-row items-center">
							<div class="flex flex-col items-center">
								<div class="text-sm">Found password: {bruteforceResult.password}</div>
							</div>
						</div>
					{:else}
						<div class="flex flex-row items-center">
							<div class="flex flex-col items-center">
								<div class="text-sm">No password found</div>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/each}
	<div class="divider">Results</div>
	{#each data?.scan?.results || [] as result}
		<div
			class="card {getResultColor(
				result.severity
			)} text-lg font-bold flex flex-col gap-3 shadow-xl grow place-content-around mt-1 mb-1"
		>
			<div class="card-body flex-col lg:flex-row">
				<div class="flex flex-row items-center gap-3 grow justify-between">
					<div class="flex flex-row items-center gap-3">
						<div class="flex flex-col grow">
							<h2 class="overflow-auto break-all">{result.message}</h2>
							<h2 class="overflow-auto break-all text-sm">
								Severity {getNameSeverity(result.severity)}
							</h2>
						</div>
					</div>
					<div class="flex flex-row items-center">
						<div class="flex flex-col items-center">
							<div class="text-sm">{result.created_at}</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/each}
</div>
