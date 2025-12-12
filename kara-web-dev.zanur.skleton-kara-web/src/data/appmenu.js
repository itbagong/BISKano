const appMenu = [
  {
    label:'My Kara',
    icon:'store-marker-outline',
    submenu: [
      [
        { label: 'Dashboard', url: '/kara'},
        { label: 'ToDos', url: '/kara'},        
      ],
      [ { label: 'Back to Xibar', url: '/app'}, ],
    ]
  },
  {
    label: 'Transaction',
    icon: 'timer-marker-outline',
    submenu: [
      [
        { label: 'Checkin / Out', url: '/' },
        { label: 'Checkin / Out - By Others', url: '/' },
        { label: 'Leave', url: '/' },
      ],
    ],
  },
  {
    label: 'Report',
    icon: 'map-clock-outline',
    submenu: [
      [
        { label: 'Attendance', url: '/' },
        { label: 'Shift Plan', url: '/' },
      ],
    ],
  },
  {
    label: 'Configuration',
    icon: 'cogs',
    submenu: [
      [
        { label: 'Work Location', url: '/' },
        { 
          label: 'Rules', 
          submenu: [
            [ 
              { label: 'General', url: '/' },
              { label: 'Rule Lines', url: '/' },
            ]
          ] 
        },
        { label: 'User Profile', url: '/' },
      ],
    ],
  },
];

export default appMenu