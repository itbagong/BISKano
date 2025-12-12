const appMenu = [
    { label: 'Home', icon: 'home', url: '/' },
    {
      label: 'Kara',
      icon: 'city',
      submenu: [
        [
          {label:'Transaction', url:'/kara/trx'}
        ],
        [
          { label: 'Location', url: '/kara/location' },
          { label: 'Rules', url: '/kara/rule' },
          { label: 'User', url: '/kara/user' },
        ],
      ],
    },
    {
      label: 'messaging',
      icon: 'messaging',
      submenu: [[{ label: 'Template messaging', url: '/msg/tpl' }]],
    },
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
          { label: 'Role', url: '/admin/role' },
          { label: 'User', url: '/admin/user' },
          { label: 'Access Grant', url: '/admin/grant' },
        ]
      ],
    },
    {label: 'Docums', icon:'file-document-outline', submenu: [
      [{label:'Docums Home', url:'/docums/index'}],
      [
        {label:'Action', url:'/general/tableview?objname=docums/action&title=Action'},
        {label:'Organization', url:'/general/tableview?objname=docums/org&title=Organization'},
        {label:'Location', url:'/general/tableview?objname=docums/orgloc&title=Organization%20location'},
        {label:'Category', url:'/general/tableview?objname=docums/category&title=Category'},
        {label:'Document Type', url:'/docums/docutype'},
      ]
      ]
    },
    {
      label: 'Labs', icon: 'school', submenu: [
          [{ label: 'Index', url: '/lab/index' }],
          [
              { label: 'Customer Group', url: '/general/tableview?objname=lab/customergroup&title=Customer%20Group' },
              { label: 'Customer', url: '/general/tableview?objname=lab/customer&title=Customer' },
              { label: 'Item', url: '/lab/item' },
          ],
          [
              { label: 'Sales', url: '/lab/sales' },
          ]
      ]  
      },
      {
        label: 'Core',
        icon: 'cog',
        submenu: [
          [
            {label:'Sequence', url:'/general/tableview?objname=admin-core/ns&title=Number%20Sequence'},
            {label:'Config', url:'/admin-core/config'},
          ],
          [
            {label:'Currency', url:'/general/tableview?objname=admin-core/currency&title=Currency'},
            {label:'Country', url:'/general/tableview?objname=admin-core/country&title=Country'},
            {label:'Payment Term', url:'/general/tableview?objname=admin-core/paymentterm&title=Payment%20Term'},
          ],
          [
            { label: 'Customer Group', url: '/general/tableview?objname=admin-core/customergroup&title=Customer%20Group'},
            { label: 'Customer', url: '/general/tableview?objname=admin-core/customer&title=Customer'},
          ],
        ],
      },
  ]

  export default appMenu