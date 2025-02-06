<script lang="ts">
	import { EMAIL_SUFFIX, AdminCoupon } from '$lib/cart';
	import { fetchAdminCoupons, updateCoupon } from '$lib/api';
	import { onMount } from 'svelte';
	import Button from '$lib/Button.svelte';
	import TrashIcon from '$lib/TrashIcon.svelte';
	import ErrorBoundary from '$lib/ErrorBoundary.svelte';

	let loading = $state(true);
	let coupons: AdminCoupon[] = $state([]);
	let error: unknown = $state();
	onMount(() => {
		fetchAdminCoupons()
			.then((x) => {
				loading = false;
				coupons = x;
			})
			.catch((e) => {
				error = e;
			});
	});

	$effect(() => {
		if (!loading && (coupons.length === 0 || coupons[coupons.length - 1].id !== null)) {
			coupons.push({
				id: null,
				enabled: false,
				public: false,
				stripe_id: '',
				stripe_desc: '(New Coupon)',
				couponCode: '',
				requirements: [],
				discount: {
					type: 'percentage',
					amount: 0
				}
			});
		}
	});

	const reqDesc = $derived(
		coupons.map((x) => {
			if (!x.requirements) return '-';
			return x.requirements
				.map<string>((req) => {
					switch (req.type) {
						case 'purchase_count':
							return `Buy ${req.amount}`;
						case 'email':
							return `Email: ${req.value}`;
					}
				})
				.join(', ');
		})
	);
	const discountDesc = $derived(
		coupons.map((x) => {
			switch (x.discount.type) {
				case 'percentage':
					return `${x.discount.amount}%`;
			}
		})
	);

	let selected = $state(-1);
	const select = (i: number) => () => {
		selected = i;
	};

	const update = async () => {
		try {
			coupons[selected] = await updateCoupon(coupons[selected]);
		} catch (e) {
			error = e;
		}
	};

	const addRequirement = () => {
		coupons[selected].requirements.push({
			type: 'purchase_count',
			amount: 0
		});
	};
	const removeRequirement = (i: number) => () => {
		coupons[selected].requirements.splice(i, 1);
	};
</script>

<table class="w-fit border border-black text-center">
	<thead>
		<tr>
			<th>#</th>
			<th>Coupon Code</th>
			<th>Stripe Desc</th>
			<th>Stripe ID</th>
			<th>Enabled</th>
			<th>Public</th>
			<th>Requirements</th>
			<th>Discounts</th>
		</tr>
	</thead>
	<tbody>
		{#each coupons as coupon, i}
			<tr class="odd:bg-gray-200" onclick={select(i)}>
				<td>{coupon.id}</td>
				<td>{coupon.couponCode}</td>
				<td>{coupon.stripe_desc}</td>
				<td>{coupon.stripe_id}</td>
				<td>{coupon.enabled ? 'YES' : ''}</td>
				<td>{coupon.public ? 'YES' : ''}</td>
				<td>{reqDesc[i]}</td>
				<td>{discountDesc[i]}</td>
			</tr>
		{/each}
	</tbody>
</table>

{#if selected !== -1}
	<div class="grid grid-cols-[auto,1fr] gap-x-2 gap-y-1">
		<span>ID</span>
		<input disabled value={coupons[selected].id} placeholder="Coupon ID (auto-generated)" />
		<span>Enabled</span>
		<input type="checkbox" bind:checked={coupons[selected].enabled} />
		<span>Auto-Apply</span>
		<input type="checkbox" bind:checked={coupons[selected].public} />
		<span>Stripe ID</span>
		<input
			disabled
			value={coupons[selected].stripe_id}
			placeholder="Coupon ID in Stripe (auto-generated)"
		/>
		<span>Coupon Code (Store UI)</span>
		<input bind:value={coupons[selected].couponCode} />
		<span>Stripe Description</span>
		<input bind:value={coupons[selected].stripe_desc} />
		<span class="header">Discount</span>
		<span>Type</span>
		<select bind:value={coupons[selected].discount.type}>
			<option value="percentage">Percentage Off</option>
		</select>
		{#if coupons[selected].discount.type === 'percentage'}
			<span>Amount Off (%)</span>
			<input bind:value={coupons[selected].discount.amount} type="number" />
		{/if}
		<span class="header">Requirements</span>
		{#each coupons[selected].requirements as req, i}
			<div class="flex items-center justify-between self-start">
				<select bind:value={req.type}>
					<option value="purchase_count">Minimum Purchase Quantity</option>
					<option value="email">Buyer Email (include {EMAIL_SUFFIX}!)</option>
				</select>
				<button onclick={removeRequirement(i)}><TrashIcon classes="size-4" /></button>
			</div>
			<div>
				{#if req.type === 'purchase_count'}
					<input bind:value={req.amount} type="number" />
				{:else if req.type === 'email'}
					<input bind:value={req.value} />
				{/if}
			</div>
		{/each}
		<span class="col-span-2 flex">
			<Button size="md" onClick={addRequirement}>Add Requirement</Button>
		</span>
		<div class="flex"><Button onClick={update}>Update Coupon</Button></div>
	</div>
{/if}

<ErrorBoundary {error} />

<style lang="postcss">
	th,
	td {
		@apply px-4 py-1;
	}
	input {
		@apply max-w-64 border border-black px-1;
	}
	input[type='checkbox'] {
		@apply justify-self-start;
		max-width: unset;
	}
	select {
		@apply max-w-64 border border-black bg-white px-1 py-1;
	}
	.header {
		@apply col-span-2 text-xl;
	}
</style>
