<script lang="ts">
	import { enhance } from '$app/forms';
	import type { ActionData } from './$types';
	import type { PageData } from './$types';

	export let data: PageData;
	export let form: ActionData;
</script>

{#if !data.user?.has_totp}
	<div class="hero min-h-screen bg-base-200">
		<div class="text-center hero-content">
			<div class="max-w-md">
				<h1 class="text-5xl font-bold">Start TOTP Registration</h1>
				<p class="py-6">
					Click the button below to begin the process of setting up Time-based One-Time Password
					(TOTP) authentication.
				</p>
				{#if form?.error}
					<div class="alert alert-error">
						{form.error}
					</div>
				{/if}
				<form use:enhance method="POST" action="?/start">
					<button type="submit" class="btn btn-primary">Start TOTP Registration</button>
				</form>
			</div>
		</div>
	</div>
{:else}
	<div class="hero min-h-screen bg-base-200">
		<div class="text-center hero-content">
			<div class="max-w-md">
				<h1 class="text-5xl font-bold">Remove TOTP</h1>
				<p class="py-6">
					You cannot start TOTP registration because you have already set up TOTP authentication.
					Enter your current TOTP token below to remove TOTP authentication.
				</p>
				{#if form?.error}
					<div class="alert alert-error">
						{form.error}
					</div>
				{/if}
				<form use:enhance method="POST" class="flex gap-2 flex-col" action="?/remove">
					<div class="form-control">
						<label class="label" for="token">
							<span class="label-text">Current Token</span>
						</label>
						<input
							type="text"
							name="token"
							placeholder="Current Token"
							class="input input-bordered"
						/>
					</div>
					<button type="submit" class="btn btn-primary">Remove TOTP</button>
				</form>
			</div>
		</div>
	</div>
{/if}
