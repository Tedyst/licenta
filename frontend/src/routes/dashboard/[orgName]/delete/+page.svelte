<script lang="ts">
	import { applyAction, enhance } from '$app/forms';
	import { invalidate, invalidateAll } from '$app/navigation';
	import type { ActionData, PageData } from './$types';

	export let form: ActionData;
	export let data: PageData;
</script>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left">
			<p class="py-6">
				Here you can delete the organization <b>{data.organization?.name}</b>. This action is
				irreversible.
			</p>
		</div>
		<div class="card shrink-0 w-full max-w-sm shadow-2xl bg-base-100">
			<div class="card-body">
				<form method="POST" use:enhance>
					<input type="hidden" name="organizationId" value={data.organization?.id} />
					{#if form?.error}
						<div class="label text-error text-xs">
							{form?.error}
						</div>
					{/if}

					<label class="label" for="submit">
						<span class="label-text">Are you sure that you want to delete the organization?</span>
					</label>
					<div class="form-control mt-6">
						<button class="btn btn-error" id="submit">Delete the organization</button>
					</div>
				</form>
			</div>
		</div>
	</div>
</div>
