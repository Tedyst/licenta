<script lang="ts">
	import { applyAction, enhance } from '$app/forms';
	import { invalidate } from '$app/navigation';
	import type { PageData, ActionData } from './$types';
	export let data: PageData;
	export let form: ActionData;
</script>

<svelte:head>
	<title>Create Project | {data.organization?.name} | Dashboard | Licenta</title>
</svelte:head>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left">
			<p class="py-6">
				Projects represent a collection of databases and secret sources. You can create multiple
				projects under an organization.
			</p>
		</div>
		<div class="card shrink-0 w-full max-w-sm shadow-2xl bg-base-100">
			<form
				class="card-body"
				method="POST"
				use:enhance
			>
				<div class="form-control">
					{#if form?.error}
						<div class="label text-error text-xs">
							{form.error}
						</div>
					{/if}
					<label class="label" for="projectName">
						<span class="label-text">Project Name</span>
					</label>
					<input
						type="text"
						id="projectName"
						placeholder="Name"
						class="lowercase input input-bordered transition-colors duration-300 ease-in-out {form?.error
							? 'wiggle input-error'
							: ''}"
						required
						name="projectName"
					/>
					<input type="hidden" name="organizationId" value={data.organization?.id} />
					<input type="hidden" name="organizationName" value={data.organization?.name} />
				</div>
				<div class="form-control mt-6">
					<button class="btn btn-primary">Create a new Project</button>
				</div>
			</form>
		</div>
	</div>
</div>
