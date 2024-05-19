<script lang="ts">
	import { applyAction, enhance } from '$app/forms';
	import { invalidate } from '$app/navigation';
	import type { ActionData } from './$types';

	export let form: ActionData;
</script>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left">
			<p class="py-6">
				Organizations represent the top level in your hierarchy. You'll be able to bundle a
				collection of projects within an organization as well as give organization-wide permissions
				to users.
			</p>
		</div>
		<div class="card shrink-0 w-full max-w-sm shadow-2xl bg-base-100">
			<form method="POST" class="card-body" use:enhance>
				<div class="form-control">
					<label class="label" for="organizationName">
						<span class="label-text">Organization Name</span>
					</label>
					<input
						type="text"
						name="organizationName"
						placeholder="Organization Name"
						class="input input-bordered transition-colors duration-300 ease-in-out {form?.error
							? 'wiggle input-error'
							: ''} lowercase placeholder-shown:normal-case"
						required
					/>
					{#if form?.error}
						<div class="label text-error text-xs">
							{form?.error}
						</div>
					{/if}
				</div>
				<div class="form-control mt-6">
					<button class="btn btn-primary">Create a new Organization</button>
				</div>
			</form>
		</div>
	</div>
</div>
