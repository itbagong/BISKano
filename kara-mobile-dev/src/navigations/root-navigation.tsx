import React from 'react';
import {createNativeStackNavigator} from '@react-navigation/native-stack';
import {
  NavigationContainer,
  createNavigationContainerRef,
} from '@react-navigation/native';
import AppStack from './app-stack';
import Login from '@screens/login';
import {useAppState} from '@overmind/index';

const navigationRef = createNavigationContainerRef();
const Stack = createNativeStackNavigator();
interface StackProps {
  navigation?: any;
  route?: any;
}

const RootNav = ({}: StackProps) => {
  const {isSignedIn} = useAppState();
  React.useEffect(() => {
    if (!isSignedIn && navigationRef.isReady()) {
      navigationRef.navigate('Login');
    }
    // console.log('roooooot', isSignedIn, navigationRef.isReady());
  }, [isSignedIn]);
  return (
    <NavigationContainer ref={navigationRef}>
      <Stack.Navigator initialRouteName="Login">
        <Stack.Screen
          name="Login"
          component={Login}
          options={{
            headerShown: false,
          }}
        />
        <Stack.Screen
          name="AppStack"
          component={AppStack}
          options={{
            headerShown: false,
            headerStyle: {
              backgroundColor: '#f4511e',
            },
          }}
        />
        {/* <Stack.Screen
          name="AppStack"
          component={AppStack}
          options={{
            headerShown: false,
            headerStyle: {
              backgroundColor: '#f4511e',
            },
          }}
        />
        <Stack.Screen
          name="TenantSelector"
          component={TenantSelector}
          options={{
            headerTitle: 'Select Tenant',
            ...options.textHeader,
          }}
        /> */}
      </Stack.Navigator>
    </NavigationContainer>
  );
};

export default RootNav;
