<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import client from '$lib/client';

	let organizationName = '';
	let error = '';

	const createOrganization = () => {
		client
			.POST('/organizations', { body: { name: organizationName.toLowerCase() } })
			.then(async (res) => {
				await invalidate('app:organizationinfo');
				if (res.data?.success) {
					await goto(`/dashboard/${res.data.organization.name}`);
				} else {
					error = res.error?.message || 'Internal server error';
				}
			})
			.catch((err) => {
				error = err.toString();
			});
	};
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
			<form
				class="card-body"
				on:submit|preventDefault={createOrganization}
				on:input={() => (error = '')}
			>
				<div class="form-control">
					<label class="label" for="organizationName">
						<span class="label-text">Organization Name</span>
					</label>
					<input
						type="text"
						id="organizationName"
						placeholder="Name"
						class="lowercase input input-bordered transition-colors duration-300 ease-in-out {error
							? 'wiggle input-error'
							: ''}"
						required
						bind:value={organizationName}
					/>
					{#if error}
						<div class="label text-error text-xs">
							{error}
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
