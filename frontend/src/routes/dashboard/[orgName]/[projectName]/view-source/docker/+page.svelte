<script lang="ts">
	import type { PageData } from './$types';
	import DockerIcon from '$lib/images/docker-icon.svg';
	import BaseListItem from '$lib/dashboard/BaseListItem.svelte';

	export let data: PageData;
</script>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left">
			<p class="py-6">
				Here you can view secrets found from this Docker image. Most of these are usually false
				positives, and are harmless.
			</p>
		</div>
		<BaseListItem
			databaseUrl={`${data.image?.docker_image}`}
			databaseIcon={DockerIcon}
			databaseType="Docker"
		/>

		{#each data?.layers || [] as layer, i}
			<div
				class={'collapse collapse-arrow ' +
					(layer.results.length ? 'bg-error text-white' : 'bg-base-300')}
			>
				<input type="radio" name="my-accordion-1" checked={i === 0} />
				<div class="collapse-title overflow-hidden">
					<div class="flex justify-between items-center">
						<div class="text-xl font-medium break-all">{layer.layer_hash}</div>
						<div class="text-base">{layer.scanned_at}</div>
					</div>
				</div>
				<div class="collapse-content gap-2 flex flex-col text-black">
					{#each layer?.results as result}
						<div class="card bg-slate-200">
							<div class="card-body">
								<div class="card-title">
									Offending Secret: {result.name}
								</div>
								<p class="text-gray-500 break-all"><b>File:</b> {result.filename}</p>
								<p class="text-gray-500 break-all"><b>Line:</b> {result.line}</p>
								<p class="text-gray-500 break-all">
									<b>Line Number:</b>
									{result.line_number}
								</p>
								<p class="text-gray-500 break-all">
									<b>Username:</b>
									{result.username}
								</p>
								<p class="text-gray-500 break-all">
									<b>Password:</b>
									{result.password}
								</p>
							</div>
						</div>
					{/each}
					{#if layer.results.length === 0}
						<div class="card bg-base-300">
							<div class="card-body">
								<div class="card-title">No secrets found in this layer.</div>
							</div>
						</div>
					{/if}
				</div>
			</div>
		{/each}
	</div>
</div>
