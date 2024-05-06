<script lang="ts">
	import type { components } from '$lib/api/v1';
	import GravatarImage from './GravatarImage.svelte';
	import TrashCan from 'svelte-material-icons/TrashCan.svelte';

	let modal: HTMLDialogElement | null = null;
	let modalUser: components['schemas']['OrganizationUser'] | null = null;
	let modalRole: 'Owner' | 'Admin' | 'Viewer' | 'None' = 'None';
	let selectElement: HTMLSelectElement;
	let addUserInput: string = '';

	export let members: components['schemas']['OrganizationUser'][];
	export let currentUser: components['schemas']['User'] | null = null;

	let currentUserRole = members
		.filter(
			(member: components['schemas']['OrganizationUser']) => member.email === currentUser?.email
		)
		?.at(0)?.role;
	const canEditOwner = currentUserRole === 'Owner';
	const canEditAdmin = currentUserRole === 'Owner' || currentUserRole === 'Admin';
	const canEditViewer = currentUserRole === 'Owner' || currentUserRole === 'Admin';
	const canEditNone = currentUserRole === 'Owner' || currentUserRole === 'Admin';

	export let editRoleAction: (role: 'Owner' | 'Admin' | 'Viewer' | 'None', userId: number) => void;
	export let addUserAction: (email: string) => void;
	export let deleteUserAction: (userId: number) => void;

	let editRole = (role: 'Owner' | 'Admin' | 'Viewer' | 'None', userId: number) => {
		editRoleAction(role, userId);
		modal?.close();
	};

	let onRoleChange = (userId: number) => (event: Event) => {
		let target = event.target as HTMLSelectElement;
		modalUser = members.filter((member) => member.id === userId)?.at(0) || null;
		modalRole = target.value as 'Owner' | 'Admin' | 'Viewer' | 'None';
		selectElement = target;
		modal?.showModal();
	};

	let onModalClose = () => {
		selectElement.value = modalUser?.role || 'None';
		modal?.close();
	};
</script>

<div class="flex flex-col flex-1">
	<div class="overflow-x-auto">
		<table class="table">
			<thead>
				<tr>
					<th>Name</th>
					<th>Email</th>
					<th>Role</th>
				</tr>
			</thead>
			<tbody>
				{#each members as member}
					<tr>
						<td>
							<div class="flex items-center gap-3">
								<div class="avatar">
									<div class="mask mask-squircle w-12 h-12">
										<GravatarImage email={member.email} size={56} />
									</div>
								</div>
								<div>
									<div>
										{member.username}
									</div>
								</div>
							</div>
						</td>
						<td>{member.email}</td>
						<td>
							<select
								class="select w-full max-w-xs"
								disabled={!canEditNone || member.email === currentUser?.email}
								on:change={onRoleChange(member.id)}
							>
								<option selected={member.role === 'Owner'} disabled={!canEditOwner}>Owner</option>
								<option selected={member.role === 'Admin'} disabled={!canEditAdmin}>Admin</option>
								<option selected={member.role === 'Viewer'} disabled={!canEditViewer}>Viewer</option
								>
								<option selected={member.role === 'None'} disabled={!canEditNone}>None</option>
							</select>
						</td>
						<td>
							<button
								type="button"
								class="mr-5 inline place-content-center text-red-500"
								on:click|preventDefault={() => deleteUserAction(member.id)}
							>
								<TrashCan size={25} />
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<form on:submit|preventDefault={() => addUserAction(addUserInput)} class="flex justify-center">
		<label class="form-control w-full max-w-xs">
			<div class="label">
				<span class="label-text">Add another user</span>
			</div>
			<div class="flex flex-row w-max">
				<input
					type="email"
					placeholder="Type here"
					class="input input-bordered w-full max-w-xs"
					bind:value={addUserInput}
				/>
				<button class="btn btn-primary">Add</button>
			</div>
		</label>
	</form>
</div>

<dialog class="modal" bind:this={modal} on:close={onModalClose}>
	<div class="modal-box">
		<h3 class="font-bold text-lg">
			Are you sure that you want to change <b>{modalUser?.username}</b>'s role to {modalRole}?
		</h3>
		<p class="py-4">Press ESC key or click the button below to close</p>
		<div class="modal-action">
			<form method="dialog">
				<button class="btn">No</button>
				<button
					class="btn bg-red-500 text-black"
					on:click|stopPropagation|preventDefault={() => editRole(modalRole, modalUser?.id || 0)}
					>Yes</button
				>
			</form>
		</div>
	</div>
</dialog>
