<script lang="ts">
	import { checkout, fetchCoupon } from '$lib/api';
	import {
		EMAIL_SUFFIX,
		type CartItem,
		type Coupon,
		applyCoupon,
		checkRequirement
	} from '$lib/cart';
	import Button from '$lib/Button.svelte';
	import Input from '$lib/Input.svelte';
	import Invoice from '$lib/Invoice.svelte';
	import ErrorBoundary from '$lib/ErrorBoundary.svelte';
	import IconX from '$lib/IconX.svelte';

	interface Props {
		cart: CartItem[];
		availableCoupons: Coupon[];
		close?: () => void;
		closeClass?: string;
	}

	let { cart = $bindable(), availableCoupons, close, closeClass }: Props = $props();

	const validateName = (name: string) => name !== '';
	const validateMatricNum = (matricNum: string) => /^[UG]\d{7}[A-Z]$/.test(matricNum);
	const validateEmail = (email: string) => /^[A-Za-z0-9]+$/.test(email);

	const findBestCoupon = (cart: CartItem[], coupons: Coupon[]) => {
		return coupons.reduce<Coupon | null>((prev, candidate) => {
			// Choose the better coupon (prev).
			if (prev && applyCoupon(cart, prev) < applyCoupon(cart, candidate)) return prev;
			// Use the candidate if we can.
			if (candidate.requirements.every((x) => checkRequirement(cart, userEmail, x)))
				return candidate;
			// Otherwise, use whatever we had previously.
			return prev;
		}, null);
	};
	const searchCoupon = async (cart: CartItem[], couponCode: string) => {
		let candidate = availableCoupons.find((x) => x.couponCode === couponCode);
		if (!candidate) {
			candidate = await fetchCoupon(couponCode);
		}
		if (candidate.requirements.every((x) => checkRequirement(cart, userEmail, x))) return candidate;
		throw new Error('Coupon requirement not matched');
	};

	const bestCoupon = $derived(findBestCoupon(cart, availableCoupons));

	let userName = $state('');
	let userMatricNumber = $state('');
	let userEmail = $state('');
	let coupon: Coupon | undefined = $state(undefined);
	let couponCode = $state('');
	let couponConfirmedInvalid = $state(false);
	$effect(() => {
		couponCode = bestCoupon?.couponCode || '';
	});
	$effect(() => {
		couponConfirmedInvalid = false;
		if (!couponCode) {
			coupon = undefined;
			return;
		}
		// Mark user email as used as it is used in the Promise which is otherwise not captured by $effect.
		userEmail;
		searchCoupon(cart, couponCode)
			.then((x) => {
				if (!x.requirements.every((req) => checkRequirement(cart, userEmail, req))) {
					throw new Error('Requirement not fulfilled.');
				}
				coupon = x;
			})
			.catch(() => {
				coupon = undefined;
				couponConfirmedInvalid = true;
			});
	});

	const [checkoutValid, checkoutTooltip] = $derived.by(() => {
		if (cart.length === 0) return [false, 'Add Items to Checkout'];
		if (!validateName(userName)) return [false, 'Name Missing'];
		if (!validateMatricNum(userMatricNumber)) return [false, 'Matric Number Required'];
		if (!validateEmail(userEmail)) return [false, 'NTU Email Required'];
		if (couponCode !== '' && !coupon) return [false, 'Coupon Invalid'];
		return [true, 'Checkout'];
	});

	let checkoutError: unknown = $state();
	const processCheckout = async () => {
		try {
			const checkoutURL = await checkout(
				cart,
				userName,
				userMatricNumber,
				userEmail + EMAIL_SUFFIX,
				coupon?.couponCode
			);
			window.location.href = checkoutURL;
		} catch (e) {
			checkoutError = e;
		}
	};

	let checkoutButton: Button;
	const clickCheckout = () => checkoutButton.click();
</script>

<div class="flex min-h-full flex-col justify-between gap-4">
	<div>
		<div class="flex justify-between">
			<h1 class="text-2xl">Cart</h1>
			<button onclick={close} class={closeClass}><IconX /></button>
		</div>
		<div class="full flex flex-col gap-2 lg:px-8">
			<div class="my-2">
				{#if cart.length === 0}
					<p class="text-center text-sm italic">The cart is currently empty.</p>
				{:else}
					<Invoice bind:items={cart} {coupon} editable />
				{/if}
			</div>
		</div>
	</div>
	<div class="flex flex-col gap-2">
		<form class="grid grid-cols-1 xl:grid-cols-2" onsubmit={clickCheckout}>
			<Input label="Name" bind:value={userName} validate={validateName} />
			<Input label="Matric Number" bind:value={userMatricNumber} validate={validateMatricNum} />
			<div class="flex items-end xl:col-span-2">
				<div class="min-w-0 flex-grow">
					<Input label="Email" bind:value={userEmail} validate={validateEmail} />
				</div>
				<span class="text-lg">{EMAIL_SUFFIX}</span>
			</div>
			<div class="xl:col-span-2">
				<Input label="Coupon Code" bind:value={couponCode} invalid={couponConfirmedInvalid} />
			</div>
			<!-- https://stackoverflow.com/questions/4196681/form-not-submitting-when-pressing-enter -->
			<input type="submit" class="hidden" />
		</form>
		<Button disabled={!checkoutValid} onClick={processCheckout} bind:this={checkoutButton}>
			{checkoutTooltip}
		</Button>
		<ErrorBoundary error={checkoutError} />
	</div>
</div>
