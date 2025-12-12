/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react/no-unstable-nested-components */
import React from 'react';
import {
  Text,
  View,
  StyleSheet,
  BackHandler,
  Image,
  TouchableWithoutFeedback,
} from 'react-native';
// import {Button, Checkbox, TextInput} from 'react-native-paper';
import {
  Button,
  CheckBox,
  Icon,
  IconElement,
  Input,
  Spinner,
} from '@ui-kitten/components';
import {Colors, Mixins, Typography} from '@utils/index';
import screenStyles from '@components/styles';
// import {showMessage} from 'react-native-flash-message';
import {useActions} from '@overmind/index';
import AsyncStorage from '@react-native-async-storage/async-storage';
import images from 'assets/images';

type Props = {
  navigation: any;
};

const Login = ({navigation}: Props) => {
  const [UserID, setUserID] = React.useState('');
  const [Password, setPassword] = React.useState('');
  const [secure, setSecure] = React.useState(true);
  const [loading, setLoading] = React.useState(false);
  const [rememberMe, setRememberMe] = React.useState(false);
  const {signIn} = useActions();

  const setRememberUser = async (value: any) => {
    try {
      const jsonValue = JSON.stringify(value);
      await AsyncStorage.setItem('@userlogin', jsonValue);
    } catch (e) {
      // saving error
    }
  };
  const getRememberUser = async () => {
    try {
      const jsonValue = await AsyncStorage.getItem('@userlogin');
      if (jsonValue != null) {
        let dataUser = JSON.parse(jsonValue);
        setRememberMe(true);
        setUserID(dataUser.userID);
        setPassword(dataUser.password);
      }
    } catch (e) {
      // error reading value
    }
  };
  const removeRememberUser = async () => {
    try {
      await AsyncStorage.removeItem('@userlogin');
    } catch (e) {
      // remove error
    }
  };
  React.useEffect(() => {
    getRememberUser();
  }, []);
  const doLogin = () => {
    if (loading) {
      return;
    }
    setLoading(true);

    signIn({
      payload: {
        CheckName: 'LoginID',
      },
      auth: {
        username: UserID,
        password: Password,
      },
    })
      .then((Data: any) => {
        global.token = Data.Token;
        if (rememberMe) {
          setRememberUser({
            userID: UserID,
            password: Password,
          });
        } else {
          removeRememberUser();
        }
        navigation.replace('AppStack');
      })
      .catch((e: any) => {
        console.log(e);
        // showMessage({
        //   type: 'danger',
        //   message: e,
        // });
        setLoading(false);
      });
  };
  const handleBackButtonClick = () => {
    BackHandler.exitApp();
    return true;
  };
  React.useEffect(() => {
    BackHandler.addEventListener('hardwareBackPress', handleBackButtonClick);
    return () => {
      BackHandler.removeEventListener(
        'hardwareBackPress',
        handleBackButtonClick,
      );
    };
  }, []);
  const SignInIcon = (props: any): IconElement => (
    <Icon {...props} name="log-in" />
  );
  const renderIcon = (props: any): React.ReactElement => (
    <TouchableWithoutFeedback onPress={() => setSecure(!secure)}>
      <Icon {...props} name={secure ? 'eye-off' : 'eye'} />
    </TouchableWithoutFeedback>
  );
  return (
    <View style={styles.container}>
      <Image
        style={{
          position: 'absolute',
          width: Mixins.scaleSize(100),
          height: Mixins.scaleSize(100),
          top: Mixins.scaleSize(-60),
          borderTopLeftRadius: Mixins.scaleSize(30),
        }}
        source={images.vector1}
      />
      <Text
        style={{
          ...Typography.headerSmSemiBold,
          color: Colors.SHADES.dark[700],
          marginBottom: Mixins.scaleSize(10),
        }}>
        Login to your account
      </Text>
      <View
        style={[
          screenStyles.column,
          styles.inputGroup,
          {marginBottom: Mixins.scaleSize(20)},
        ]}>
        <Input
          label="Login ID"
          placeholder="Login ID"
          value={UserID}
          onChangeText={nextValue => setUserID(nextValue)}
        />
      </View>
      <View style={[screenStyles.column, styles.inputGroup]}>
        <Input
          value={Password}
          label="Password"
          placeholder="Password"
          accessoryRight={renderIcon}
          secureTextEntry={secure}
          onChangeText={nextValue => setPassword(nextValue)}
        />
      </View>
      <View style={[screenStyles.row, styles.checkBoxGroup]}>
        <CheckBox
          checked={rememberMe}
          style={{marginRight: Mixins.scaleSize(10)}}
          onChange={nextChecked => {
            setRememberMe(nextChecked);
          }}
        />
        <Text style={styles.checkBoxText}>Remember me</Text>
      </View>
      <View
        style={{
          flexDirection: 'row',
          justifyContent: 'center',
        }}>
        <Button
          style={{flex: 1}}
          size="large"
          status="warning"
          onPress={doLogin}
          accessoryLeft={(props: any) => {
            return loading ? (
              <Spinner size="small" status="basic" />
            ) : (
              SignInIcon(props)
            );
          }}>
          <Text style={styles.buttonLabel}>
            {loading ? 'Loading' : 'LOGIN'}
          </Text>
        </Button>
      </View>
    </View>
  );
};

export default Login;

const styles = StyleSheet.create({
  container: {
    flex: 1,
    position: 'relative',
    marginTop: Mixins.scaleSize(60),
    marginBottom: Mixins.scaleSize(35),
    paddingTop: Mixins.scaleSize(20),
    paddingHorizontal: Mixins.scaleSize(20),
  },
  inputGroup: {
    marginBottom: Mixins.scaleSize(10),
  },
  checkBoxGroup: {
    marginVertical: Mixins.scaleSize(10),
    alignItems: 'center',
  },
  checkBoxText: {
    color: Colors.SHADES.dark[900],
    ...Typography.textLg,
  },
  inputIcon: {
    fontSize: Mixins.scaleFont(24),
    paddingTop: Mixins.scaleSize(22),
    paddingRight: Mixins.scaleSize(0),
    color: Colors.SHADES.dark[900],
  },
  inputLabel: {
    color: Colors.SHADES.dark[900],
    ...Typography.textLg,
  },
  input: {
    backgroundColor: 'transparent',
    ...Typography.textLg,
    marginBottom: -2,
  },
  buttonLabel: {
    fontSize: Mixins.scaleFont(18),
    textTransform: 'capitalize',
  },
  textReset: {
    color: Colors.PRIMARY.greenLight,
    ...Typography.textLgPlus,
  },
  indicator: {
    justifyContent: 'center',
    alignItems: 'center',
  },
});
