import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url, parent }) => {
	const parentData = await parent();

	const userId = url.searchParams.get('userId') || '0';

	const editedUser = parentData?.organization?.members.find((user) => user.id === +userId);

	const currentUserRole = parentData?.organization?.members
		.filter((member) => member.email === parentData.user.email)
		?.at(0)?.role;
	const canEditOwner = currentUserRole === 'Owner';
	const canEditAdmin = currentUserRole === 'Owner' || currentUserRole === 'Admin';
	const canEditViewer = currentUserRole === 'Owner' || currentUserRole === 'Admin';
	const canEditNone = currentUserRole === 'Owner' || currentUserRole === 'Admin';

	return { editedUser, canEditOwner, canEditAdmin, canEditViewer, canEditNone };
};
