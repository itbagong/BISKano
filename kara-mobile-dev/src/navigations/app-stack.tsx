/* eslint-disable react/no-unstable-nested-components */
/* eslint-disable react-native/no-inline-styles */
import React from 'react';
import {Text, TouchableOpacity, View, StyleSheet} from 'react-native';
import {createNativeStackNavigator} from '@react-navigation/native-stack';
import {createBottomTabNavigator} from '@react-navigation/bottom-tabs';
import {
  Avatar,
  BottomNavigation,
  BottomNavigationTab,
  IconElement,
} from '@ui-kitten/components';
import HomeScreen from '@screens/home';
import History from '@screens/history';
// import Leave from '@screens/leave';
import Routine from '@screens/routine-list';
import RoutineDetails from '@screens/routine-details';
import RoutineAssetDetails from '@screens/routine-assets-details';
import RoutineDetailsAssetChecklist from '@screens/routine-details-asset-checklist';
import Approval from '@screens/approval';
import ApprovalDetail from '@screens/approval-detail';
import ApprovalPreview from '@screens/approval-preview';
import UserProfile from '@screens/user-profile';
const {Navigator, Screen} = createBottomTabNavigator();
import {default as theme} from '../custom-theme.json'; // <-- Import app theme
import {Mixins, Colors, Typography} from 'utils';
import {useAppState} from '@overmind/index';
import images from 'assets/images';
import Feather from 'react-native-vector-icons/Feather';
import {
  ClipboardTick,
  Home,
  Notepad2,
  Refresh2,
  UserOctagon,
} from 'iconsax-react-native';

const HomeIcon = (props: any): IconElement => (
  <Home size="32" color={props.style.tintColor} variant="TwoTone" />
);

const HistoryIcon = (props: any): IconElement => (
  <Refresh2 size="32" color={props.style.tintColor} variant="TwoTone" />
);
// const LeaveIcon = (props: any): IconElement => (
//   <Icon {...props} name="log-out-outline" />
// );
const UserIcon = (props: any): IconElement => (
  <UserOctagon size="32" color={props.style.tintColor} variant="TwoTone" />
);
const ApprovalIcon = (props: any): IconElement => (
  <ClipboardTick size="32" color={props.style.tintColor} variant="TwoTone" />
);
const RoutineIcon = (props: any): IconElement => (
  <Notepad2 size="32" color={props.style.tintColor} variant="TwoTone" />
);
const BottomTabBar = ({navigation, state}: any) => (
  <BottomNavigation
    selectedIndex={state.index}
    onSelect={index => navigation.navigate(state.routeNames[index])}>
    <BottomNavigationTab icon={HomeIcon} />
    <BottomNavigationTab icon={RoutineIcon} />
    <BottomNavigationTab icon={HistoryIcon} />
    <BottomNavigationTab icon={ApprovalIcon} />
    <BottomNavigationTab icon={UserIcon} />
  </BottomNavigation>
);

const TabNavigator = () => {
  const {userInfo, notifications} = useAppState();
  const [text, setText] = React.useState('Good morning');

  React.useEffect(() => {
    const today = new Date();
    const curHr = today.getHours();

    if (curHr < 12) {
      setText('Good morning');
    } else if (curHr < 18) {
      setText('Good afternoon');
    } else {
      setText('Good evening');
    }
    return () => {};
  }, []);

  return (
    <Navigator tabBar={props => <BottomTabBar {...props} />}>
      <Screen
        name="Home"
        component={HomeScreen}
        options={{
          headerTitle: '',
          headerRight: () => (
            <TouchableOpacity onPress={() => {}} style={styles.containerNotif}>
              <View style={styles.iconNotifContainer}>
                <Feather
                  name="bell"
                  size={Mixins.scaleSize(20)}
                  color={Colors.SHADES.gray[800]}
                />
                {notifications.length > 0 && (
                  <View style={styles.badgeContainer}>
                    <Text style={styles.badge}>{notifications.length}</Text>
                  </View>
                )}
              </View>
            </TouchableOpacity>
          ),
          headerLeft: () => {
            return (
              <View
                style={{
                  flex: 1,
                  flexDirection: 'row',
                  paddingHorizontal: Mixins.scaleSize(10),
                  marginTop: Mixins.scaleSize(10),
                  gap: 10,
                  alignItems: 'center',
                }}>
                <Avatar size="medium" source={images.person} />
                <View>
                  <Text style={{...Typography.textMdPlus, color: Colors.BLACK}}>
                    {text}
                  </Text>
                  <Text style={{...Typography.textMdPlus, color: Colors.BLACK}}>
                    {userInfo?.DisplayName}
                  </Text>
                </View>
              </View>
            );
          },
          headerStyle: {
            backgroundColor: Colors.SHADES.gray[100],
          },
          headerTintColor: Colors.SHADES.dark[900],
        }}
      />
      <Screen
        name="Routine"
        component={Routine}
        options={{
          headerTitle: 'P2H',
        }}
      />
      <Screen
        name="History"
        component={History}
        options={{
          headerTitle: 'Attendance history',
          headerTitleAlign: 'center',
        }}
      />
      {/* <Screen
        name="Leave"
        component={Leave}
        options={{
          headerTitle: 'Leave',
        }}
      /> */}
      <Screen
        name="Approval"
        component={Approval}
        options={{
          headerTitle: 'Approval',
          headerTitleAlign: 'center',
        }}
      />
      <Screen
        name="UserProfile"
        component={UserProfile}
        options={{
          headerTitle: 'User profile',
          headerTitleAlign: 'center',
        }}
      />
    </Navigator>
  );
};

const AppStackNav = createNativeStackNavigator();
export const AppNavigator = () => (
  <AppStackNav.Navigator initialRouteName="BottomTabNav">
    <AppStackNav.Screen
      name="BottomTabNav"
      component={TabNavigator}
      options={{
        headerShown: false,
      }}
    />
    <AppStackNav.Screen
      name="RoutineDetails"
      component={RoutineDetails}
      options={{
        headerTitle: 'Routine Details',
        headerTitleAlign: 'center',
      }}
    />
    <AppStackNav.Screen
      name="RoutineAssetDetails"
      component={RoutineAssetDetails}
      options={{
        headerTitle: 'Asset Details',
        headerTitleAlign: 'center',
      }}
    />
    <AppStackNav.Screen
      name="RoutineDetailsAssetChecklist"
      component={RoutineDetailsAssetChecklist}
      options={{
        headerTitle: '',
        headerTitleAlign: 'center',
      }}
    />
    <AppStackNav.Screen
      name="ApprovalDetail"
      component={ApprovalDetail}
      options={{
        headerTitle: '',
        headerTitleAlign: 'center',
      }}
    />
    <AppStackNav.Screen
      name="ApprovalPreview"
      component={ApprovalPreview}
      options={{
        headerTitle: 'Preview',
        headerTitleAlign: 'center',
      }}
    />
  </AppStackNav.Navigator>
);

export default AppNavigator;

const styles = StyleSheet.create({
  containerNotif: {
    position: 'relative',
    marginRight: Mixins.scaleSize(14),
    borderRadius: 100,
    borderColor: Colors.SHADES.gray[200],
    borderWidth: 1,
    // width: Mixins.scaleSize(40),
    // height: Mixins.scaleSize(40),
    padding: Mixins.scaleSize(7),
    backgroundColor: Colors.WHITE,
  },
  iconNotifContainer: {
    position: 'relative',
  },
  iconNotif: {
    fontSize: 24,
  },
  badgeContainer: {
    position: 'absolute',
    top: -5,
    right: -8,
    minWidth: Mixins.scaleSize(15),
    height: Mixins.scaleSize(15),
    borderRadius: 10,
    backgroundColor: theme['color-primary-500'],
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 1,
  },
  badge: {
    color: Colors.BLACK,
    ...Typography.textSm,
  },
});
