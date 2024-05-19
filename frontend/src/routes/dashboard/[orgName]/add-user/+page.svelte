<script lang="ts">
	import { enhance } from '$app/forms';
	import type { ActionData, PageData } from './$types';

	export let form: ActionData;
	export let data: PageData;
</script>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left">
			<p class="py-6">
				Here you can add a new user in the organization <b>{data.organization?.name}</b>.
			</p>
		</div>
		<div class="card shrink-0 w-full max-w-sm shadow-2xl bg-base-100">
			<div class="card-body">
				<form method="POST" use:enhance action="?/editRole">
					<input type="hidden" name="organizationId" value={data.organization?.id} />
					{#if form?.error}
						<div class="label text-error text-xs">
							{form?.error}
						</div>
					{/if}

					<div class="form-control">
						<label class="label" for="email">
							<span class="label-text">Email</span>
						</label>
						<input
							type="email"
							id="email"
							placeholder="Email"
							class="input input-bordered transition-colors duration-300 ease-in-out {form?.error
								? 'wiggle input-error'
								: ''}"
							required
							name="email"
						/>

						<label class="label" for="organizationName">
							<span class="label-text">Select Role</span>
						</label>
						<select class="select select-bordered w-full max-w-xs" id="role" name="role">
							<option selected={data?.editedUser?.role === 'Owner'} disabled={!data?.canEditOwner}>
								Owner
							</option>
							<option selected={data?.editedUser?.role === 'Admin'} disabled={!data?.canEditAdmin}>
								Admin
							</option>
							<option
								selected={data?.editedUser?.role === 'Viewer'}
								disabled={!data?.canEditViewer}
							>
								Viewer
							</option>
							<option selected={data?.editedUser?.role === 'None'} disabled={!data?.canEditNone}>
								None
							</option>
						</select>
					</div>
					<div class="form-control mt-6">
						<button class="btn btn-primary">Add the user</button>
					</div>
				</form>
			</div>
		</div>
	</div>
</div>
