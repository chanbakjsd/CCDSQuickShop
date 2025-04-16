<script lang="ts">
	import { onMount } from 'svelte'

	import { permCheck } from '$lib/api'
	import ErrorBoundary from '$lib/ErrorBoundary.svelte'
	import Header from '$lib/Header.svelte'
	import Options from '$lib/Options.svelte'
	import CouponEdit from './CouponEdit.svelte'
	import ClosuresEdit from './ClosuresEdit.svelte'
	import MerchEdit from './MerchEdit.svelte'
	import UsersEdit from './UsersEdit.svelte'
	import OrderCollection from './OrderCollection.svelte'
	import OrderSummary from './OrderSummary.svelte'

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
		{ text: 'Merch' },
		{ text: 'Coupons' },
		{ text: 'Admin Users' },
		{ text: 'Order Collection' },
		{ text: 'Unfulfilled Order Summary' }
	]
	let selected: string | undefined = $state(undefined)

	let orderCollection: OrderCollection | undefined = $state(undefined)
	const searchOrder = (value: string) => {
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
		{:else if selected === 'Merch'}
			<MerchEdit />
		{:else if selected === 'Admin Users'}
			<UsersEdit />
		{:else if selected === 'Order Collection'}
			<OrderCollection bind:this={orderCollection} />
		{:else if selected === 'Coupons'}
			<CouponEdit />
		{:else if selected === 'Unfulfilled Order Summary'}
			<OrderSummary {searchOrder} />
		{/if}
	</ErrorBoundary>
</div>
