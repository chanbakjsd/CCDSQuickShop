<script lang="ts">
	import { cancelOrder, collectOrder, listOrders, type Order } from '$lib/api';
	import Button from '$lib/Button.svelte';
	import Input from '$lib/Input.svelte';
	import OrderPreview from '../orders/[id]/OrderPreview.svelte';

	let orders: Order[] = [];
	let orderInput = '';
	let includeCancelled = false;
	const searchOrders = async () => {
		orders = await listOrders(orderInput, includeCancelled);
	};
	const markCollect = (orderID: string) => async () => {
		await collectOrder(orderID);
		orders = orders.map((x) => ({
			...x,
			collectionTime: x.id === orderID ? new Date() : x.collectionTime
		}));
	};
	const markCancel = (orderID: string) => async () => {
		await cancelOrder(orderID);
		orders = orders.map((x) => ({
			...x,
			cancelled: x.id === orderID ? true : x.cancelled
		}));
	};
</script>

<div class="flex items-center gap-2">
	<form on:submit={searchOrders}>
		<Input label="Matric Number/Order ID/NTU Email" bind:value={orderInput} />
	</form>
	<label class="flex items-center gap-2 text-lg">
		<input type="checkbox" bind:checked={includeCancelled} />
		Include Cancelled
	</label>
	<Button size="md" on:click={searchOrders}>Search</Button>
</div>
<div class="flex flex-col gap-4">
	{#each orders as order}
		<OrderPreview {order} collect={markCollect(order.id)} cancel={markCancel(order.id)} />
	{/each}
</div>
