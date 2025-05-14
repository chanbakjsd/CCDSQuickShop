<script lang="ts">
	import { onMount } from 'svelte'
	import api from '$lib/api'
	import Button from '$lib/Button.svelte'
	import Icon from '$lib/icon/Icon.svelte'
	import ErrorBoundary from '$lib/ErrorBoundary.svelte'

	let error: unknown = $state()
	let adminEmail = $state('')
	let admins: string[] = $state([])
	onMount(() => {
		api.admin.users
			.list()
			.then((x) => {
				admins = x
			})
			.catch((e) => {
				error = e
			})
	})

	const addAdminEmail = async () => {
		try {
			await api.admin.users.add(adminEmail)
			admins = [...admins, adminEmail]
			adminEmail = ''
		} catch (e) {
			error = e
		}
	}

	const deleteAdminEmail = (email: string) => async () => {
		try {
			await api.admin.users.remove(email)
			admins = admins.filter((x) => x !== email)
		} catch (e) {
			error = e
		}
	}
</script>

<div class="flex flex-col gap-4 p-4">
	<ErrorBoundary {error} />

	<div class="flex items-center gap-1">
		<form onsubmit={addAdminEmail}>
			<input placeholder="Email of New Admin" bind:value={adminEmail} />
		</form>
		<Button size="md" onClick={addAdminEmail}>Add</Button>
	</div>

	<div class="admin-list">
		{#each admins as admin}
			{admin}
			<button onclick={deleteAdminEmail(admin)}>
				<Icon name="trash" />
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
