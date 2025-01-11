<script lang="ts">
	import { onMount } from 'svelte';
	import { addUser, listUsers, deleteUser } from '$lib/api';
	import Button from '$lib/Button.svelte';
	import TrashIcon from '$lib/TrashIcon.svelte';

	let adminEmail = '';
	let admins: string[] = [];
	onMount(() => {
		listUsers().then((x) => {
			admins = x;
		});
	});

	const addAdminEmail = async () => {
		await addUser(adminEmail);
		admins = [...admins, adminEmail];
		adminEmail = '';
	};

	const deleteAdminEmail = (email: string) => async () => {
		await deleteUser(email);
		admins = admins.filter((x) => x !== email);
	};
</script>

<div class="flex flex-col gap-4 p-4">
	<div class="flex items-center gap-1">
		<form on:submit={addAdminEmail}>
			<input placeholder="Email of New Admin" bind:value={adminEmail} />
		</form>
		<Button size="md" on:click={addAdminEmail}>Add</Button>
	</div>

	<div class="admin-list">
		{#each admins as admin}
			{admin}
			<button on:click={deleteAdminEmail(admin)}>
				<TrashIcon />
			</button>
		{/each}
	</div>
</div>

<style lang="postcss">
	input {
		@apply border border-black px-2;
	}
	.admin-list {
		@apply grid w-fit;
		grid-template-columns: auto auto;
	}
</style>
