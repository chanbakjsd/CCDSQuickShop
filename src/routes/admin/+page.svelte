<script lang="ts">
	import { onMount } from 'svelte'

	import { permCheck } from '$lib/api'
	import ErrorBoundary from '$lib/ErrorBoundary.svelte'
	import Header from '$lib/Header.svelte'
	import Options from '$lib/Options.svelte'
	import ClosuresEdit from './ClosuresEdit.svelte'
	import SalePeriodEdit from './SalePeriodEdit.svelte'
	import OrderCollection from './OrderCollection.svelte'
	import UsersEdit from './UsersEdit.svelte'

	let err: unknown = $state()
	onMount(() => {
		try {
			permCheck()
		} catch (e) {
			err = e
		}
	})

	let options = [
		{ text: 'Store Closures' },
		{ text: 'Admin Users' },
		{ text: 'Storefront Management' },
		{ text: 'Order Collection' }
	]
	let selected: string | undefined = $state(undefined)

	let orderCollection: OrderCollection | undefined = $state(undefined)
	const searchOrder = (value: string) => {
		if (!value) return
		selected = 'Order Collection'
		setTimeout(() => orderCollection?.search(value))
	}
</script>

<div class="flex flex-col gap-4 p-4">
	<Header admin cls="bg-white" />
	<ErrorBoundary error={err}>
		<Options {options} bind:value={selected} />
		{#if selected === 'Store Closures'}
			<ClosuresEdit />
		{:else if selected === 'Admin Users'}
			<UsersEdit />
		{:else if selected === 'Storefront Management'}
			<SalePeriodEdit {searchOrder} />
		{:else if selected === 'Order Collection'}
			<OrderCollection bind:this={orderCollection} />
		{/if}
	</ErrorBoundary>
</div>
