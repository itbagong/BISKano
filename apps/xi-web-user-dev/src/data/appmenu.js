const appMenu = [
    {label:'Home', icon:'home', url:'/'},
    {
        label: 'Admin',
        icon: 'office-building-cog',
        submenu: [
          [
            { label: 'Feature Category', url: '/admin/featurecategory' },
            { label: 'Feature', url: '/admin/feature' },
          ],
          [
            { label: 'Tenant', url: '/admin/tenant' },
            { label: 'Tenant Request', url: '/me/join-tenant'},
            { label: 'Tenant Join Review', url: '/admin/tenantjoin'},
            { label: 'Role', url: '/admin/role' },
            { label: 'User', url: '/admin/user' },
            { label: 'Access Grant', url: '/admin/grant' },
          ],
          [
            { label: 'Application', url: '/admin/app' },
            { label: 'Tenant Application', url: '/general/tableview?objname=/admin/tenantapp&title=Tenant Application' },
            { label:'Message Template', url: '/admin/msgtpl'},
          ]
        ],
      },
];

export default appMenu