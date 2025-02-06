<script lang="ts">
	import { cancelOrder, collectOrder, listOrders, type Order } from '$lib/api';
	import Button from '$lib/Button.svelte';
	import ErrorBoundary from '$lib/ErrorBoundary.svelte';
	import Input from '$lib/Input.svelte';
	import OrderPreview from '../orders/[id]/OrderPreview.svelte';

	let orders: Order[] = $state([]);
	let orderInput = $state('');
	let includeCancelled = $state(false);
	let emptyResponse = $state(false);

	let error: unknown = $state();
	const searchOrders = async () => {
		try {
			orders = await listOrders(orderInput, includeCancelled);
			emptyResponse = orders.length === 0;
		} catch (e) {
			error = e;
		}
	};
	const markCollect = (orderID: string) => async () => {
		try {
			await collectOrder(orderID);
			orders = orders.map((x) => ({
				...x,
				collectionTime: x.id === orderID ? new Date() : x.collectionTime
			}));
		} catch (e) {
			error = e;
		}
	};
	const markCancel = (orderID: string) => async () => {
		try {
			await cancelOrder(orderID);
			orders = orders.map((x) => ({
				...x,
				cancelled: x.id === orderID ? true : x.cancelled
			}));
		} catch (e) {
			error = e;
		}
	};
</script>

<div class="flex items-center gap-2">
	<form onsubmit={searchOrders}>
		<Input label="Matric Number/Order ID/NTU Email" bind:value={orderInput} />
	</form>
	<label class="flex items-center gap-2 text-lg">
		<input type="checkbox" bind:checked={includeCancelled} />
		Include Cancelled
	</label>
	<Button size="md" onClick={searchOrders}>Search</Button>
</div>
<ErrorBoundary {error} />
<div class="flex flex-col gap-4">
	{#if emptyResponse}No orders matched.{/if}
	{#each orders as order}
		<OrderPreview {order} collect={markCollect(order.id)} cancel={markCancel(order.id)} />
	{/each}
</div>
