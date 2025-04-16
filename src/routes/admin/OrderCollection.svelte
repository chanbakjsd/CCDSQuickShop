<script lang="ts">
	import { onMount } from 'svelte'
	import { cancelOrder, collectOrder, listOrders, type Order } from '$lib/api'
	import Button from '$lib/Button.svelte'
	import ErrorBoundary from '$lib/ErrorBoundary.svelte'
	import Input from '$lib/Input.svelte'
	import OrderPreview from '../orders/[id]/OrderPreview.svelte'

	let orders: Order[] = $state([])
	let orderInput = $state('')
	let includeCancelled = $state(false)
	let emptyResponse = $state(false)

	let error: unknown = $state()
	let searchButton: Button
	export const search = (keyword: string) => {
		orderInput = keyword
		searchButton.click()
	}

	let pastSearches: string[] = $state([])
	onMount(() => {
		try {
			pastSearches = JSON.parse(window.localStorage.getItem('adminSearchHistory') || '[]')
		} catch (e) {}
	})

	const searchOrders = async () => {
		try {
			pastSearches = [orderInput, ...pastSearches]
			// Remove duplicate entries.
			pastSearches = pastSearches.filter((x, i) => pastSearches.indexOf(x) === i)
			if (pastSearches.length > 5) {
				pastSearches = pastSearches.slice(1)
			}
			window.localStorage.setItem('adminSearchHistory', JSON.stringify(pastSearches))
			orders = await listOrders(orderInput, { includeCancelled, allowFromItem: true })
			emptyResponse = orders.length === 0
		} catch (e) {
			error = e
		}
	}
	const markCollect = (orderID: string) => async () => {
		try {
			await collectOrder(orderID)
			orders = orders.map((x) => ({
				...x,
				collectionTime: x.id === orderID ? new Date() : x.collectionTime
			}))
		} catch (e) {
			error = e
		}
	}
	const markCancel = (orderID: string) => async () => {
		try {
			await cancelOrder(orderID)
			orders = orders.map((x) => ({
				...x,
				cancelled: x.id === orderID ? true : x.cancelled
			}))
		} catch (e) {
			error = e
		}
	}
</script>

<div class="flex items-center gap-2">
	<form onsubmit={searchOrders}>
		<Input label="Matric Number/Order ID/NTU Email" bind:value={orderInput} />
	</form>
	<label class="flex items-center gap-2 text-lg">
		<input type="checkbox" bind:checked={includeCancelled} />
		Include Cancelled
	</label>
	<Button size="md" onClick={searchOrders} bind:this={searchButton}>Search</Button>
</div>
<div class="flex flex-wrap gap-4">
	{#if pastSearches.length > 0}
		Past Search:
		{#each pastSearches as keyword}
			<button
				class="max-w-72 overflow-hidden text-ellipsis text-nowrap text-blue-800 underline"
				onclick={() => search(keyword)}
			>
				{keyword.replaceAll(',', ' ')}
			</button>
		{/each}
	{/if}
</div>
<ErrorBoundary {error} />
<div class="flex flex-col gap-4">
	{#if emptyResponse}No orders matched.{/if}
	{#each orders as order (order.id)}
		<OrderPreview {order} collect={markCollect(order.id)} cancel={markCancel(order.id)} />
	{/each}
</div>
