<script lang="ts">
	import type { components } from '../../lib/api/v1';

	export let organizations: components['schemas']['OrganizationProjects'][];
</script>

{#if organizations}
	{#each organizations as organization, i}
		<div class="collapse bg-base-200">
			<input type="radio" name="accordion" checked={i == 1} />
			<div class="collapse-title text-xl font-medium">{organization.name}</div>
			<div class="collapse-content">
				{#each organization.projects as project}
					<li><a>{project.name}</a></li>
				{/each}
			</div>
		</div>
	{/each}
{/if}
