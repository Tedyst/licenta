<script lang="ts">
	export let organizations: components['schemas']['Organization'][] = [];

	import { page } from '$app/stores';
	import type { components } from '$lib/api/v1';

	function shouldBeChecked(i: number, orgName: string) {
		if ($page.data?.orgName) return $page.data.orgName === orgName;
		return i == 0;
	}
</script>

{#if organizations}
	{#if organizations.length === 0}
		<div class="text-center text-md font-medium">No organizations found</div>
	{/if}
	{#each organizations as organization, i}
		{#if i !== 0}
			<li class="divider m-0 flex-nowrap shrink-0 opacity-100 bg-inherit" />
		{/if}
		<li class="p-0">
			<div class="collapse collapse-arrow bg-base-300 hover:bg-base-200 p-0">
				<input
					type="radio"
					name="accordion"
					checked={shouldBeChecked(i, organization.name)}
					aria-label={organization.name}
				/>
				<div class="collapse-title text-xl font-medium w-44">
					{organization.name}
				</div>
				<div class="collapse-content">
					<ul class="m-0 p-0 before:hidden">
						<li><a href="/dashboard/{organization.name}">Information</a></li>
						<li class="divider m-0 flex-nowrap shrink-0 opacity-100 bg-inherit" />
					</ul>
					{#each organization.projects as project}
						<ul class="list-none before:hidden m-0 p-0">
							<li><a href="/dashboard/{organization.name}/{project.name}">{project.name}</a></li>
						</ul>
					{/each}
					{#if organization.projects.length === 0}
						<div class="text-center text-md font-medium">No projects found</div>
					{/if}
				</div>
			</div>
		</li>
	{/each}
{:else}
	<div class="skeleton h-20 w-full" />
	<div class="skeleton ml-4 h-10 max-w-full mt-2" />
	<div class="skeleton ml-4 h-10 max-w-full mt-2" />
	<div class="skeleton ml-4 h-10 max-w-full mt-2" />
	<div class="skeleton ml-4 h-10 max-w-full mt-2" />
{/if}
