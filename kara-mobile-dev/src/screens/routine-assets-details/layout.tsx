/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable react-native/no-inline-styles */
import {
  ActivityIndicator,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import React from 'react';
import {
  TabView,
  TabBar,
  SceneRendererProps,
  NavigationState,
} from 'react-native-tab-view';
import container from 'components/container';
import {Colors, Mixins, Typography} from 'utils';
import TabAttachment from './tab-attachment';
import TabChecklist from './tab-checklist';
import {useActions, useAppState} from '@overmind/index';
import {default as theme} from '../../custom-theme.json';
import {Icon} from '@ui-kitten/components/ui';
import {ALERT_TYPE, Toast} from 'react-native-alert-notification';

type Props = {
  route: any;
  navigation: any;
};
type State = NavigationState<{
  key: string;
  title: string;
}>;

const RoutineAssetDetails = (props: Props) => {
  const {route, navigation} = props;
  const attachRef: any = React.useRef(null);
  const [selectedRoutineAsset, setSelectedRoutineAsset] = React.useState({});
  const {saveRoutineChecklist, saveRoutineDetails, routineRequest} =
    useActions();
  const [disabledRequest, setDisabledRequest] = React.useState(true);
  const [loadingSave, setLoadingSave] = React.useState(false);
  const [loadingReq, setLoadingReq] = React.useState(false);
  const {routineChecklist} = useAppState();
  // const shouldLoadComponent = (index: any): boolean => index === selectedIndex;
  const [index, setIndex] = React.useState(0);
  const [routes] = React.useState([
    {key: 'checklist', title: 'Checklist'},
    {key: 'attachment', title: 'Attachment'},
  ]);

  React.useEffect(() => {
    setSelectedRoutineAsset(props.route.params.routineAsset);
    return () => {};
  }, [route]);

  const checkChecklist = () => {
    let status = 'NotCheckedYet';
    let checklists: any[] = [];
    routineChecklist.RoutineChecklistCategories.map((item: any) => {
      checklists = [...checklists, ...item.RoutineChecklistDetails];
    });
    if (checklists.find((o: any) => o.Status === '')) {
      status = 'NotCheckedYet';
    } else if (checklists.find((o: any) => o.Status === 'Damaged')) {
      status = 'NeedRepair';
    } else if (checklists.find((o: any) => o.Status === 'Normal')) {
      status = 'RunningWell';
    }
    return status;
  };

  const onCheckAttach = () => {
    if (attachRef.current) {
      attachRef.current?.onWriteAsset();
      attachRef.current?.onDelete();
    }
  };
  const onSave = () => {
    const status = checkChecklist();
    if (status === 'NotCheckedYet') {
      return Toast.show({
        type: ALERT_TYPE.DANGER,
        title: 'Error',
        textBody:
          'Oops! It seems that some status on the checklist remain unchecked. Please complete all items before proceeding further.',
      });
    }
    const payload = {...selectedRoutineAsset, StatusCondition: status};
    setLoadingSave(true);
    Promise.all([
      saveRoutineChecklist({
        ...routineChecklist,
        RoutineChecklist: {
          ...routineChecklist.RoutineChecklist,
          KmToday: parseFloat(routineChecklist.RoutineChecklist.KmToday),
        },
      }),
      saveRoutineDetails(payload),
      onCheckAttach(),
    ])
      .then(() => {
        Toast.show({
          type: ALERT_TYPE.SUCCESS,
          title: 'Success',
          textBody: 'Your data has been saved!',
        });
        setTimeout(() => {
          navigation.goBack();
        }, 100);
      })
      .finally(() => setLoadingSave(false));
  };

  const onRequest = () => {
    const status = checkChecklist();
    if (status === 'NotCheckedYet') {
      return Toast.show({
        type: ALERT_TYPE.DANGER,
        title: 'Error',
        textBody:
          'Oops! It seems that some status on the checklist remain unchecked. Please complete all items before proceeding further.',
      });
    }
    const payload = {...selectedRoutineAsset, StatusCondition: status};
    setLoadingReq(true);
    Promise.all([
      saveRoutineChecklist({
        ...routineChecklist,
        RoutineChecklist: {
          ...routineChecklist.RoutineChecklist,
          KmToday: parseFloat(routineChecklist.RoutineChecklist.KmToday),
        },
      }),
      saveRoutineDetails(payload),
      onCheckAttach(),
    ])
      .then(async () => {
        const payloadReq = {
          RoutineChecklistID: routineChecklist.RoutineChecklist._id,
          RoutineID: props.route.params.routine._id,
          EquipmentNo: props.route.params.routineAsset.AssetID,
          Kilometers: parseFloat(routineChecklist.RoutineChecklist.KmToday),
          Description: '',
          Site: props.route.params.routine.SiteID,
          ...routineChecklist,
        };
        await routineRequest(payloadReq).then(() => {
          Toast.show({
            type: ALERT_TYPE.SUCCESS,
            title: 'Success',
            textBody: 'Your request has been sent!',
          });
          navigation.goBack();
        });
      })
      .finally(() => setLoadingReq(false));
  };

  const isChecklistNotDamaged = () => {
    let checklists: any[] = [];
    routineChecklist.RoutineChecklistCategories.map((item: any) => {
      checklists = [...checklists, ...item.RoutineChecklistDetails];
    });
    if (checklists.find(o => o.Status === 'Damaged')) {
      return false;
    }
    return true;
  };
  React.useEffect(() => {
    const damaged = isChecklistNotDamaged();
    if (!damaged && !routineChecklist.RoutineChecklist.IsAlreadyRequest) {
      setDisabledRequest(false);
    } else {
      setDisabledRequest(true);
    }
    return () => {};
  }, [routineChecklist?.RoutineChecklistCategories]);
  const renderScene = (cfg: any) => {
    switch (cfg.route.key) {
      case 'checklist':
        return (
          <TabChecklist
            routine={props.route.params.routine}
            routineAsset={selectedRoutineAsset}
            navigation={navigation}
            isReadOnly={route?.params?.isReadOnly}
          />
        );
      case 'attachment':
        return (
          <TabAttachment
            ref={attachRef}
            isReadOnly={route?.params?.isReadOnly}
          />
        );
      default:
        break;
    }
  };
  const renderTabBar = (
    propss: SceneRendererProps & {navigationState: State},
  ) => (
    <TabBar
      {...propss}
      scrollEnabled
      indicatorStyle={styles.indicator}
      style={styles.tabbar}
      tabStyle={styles.tab}
      labelStyle={styles.label}
      activeColor={theme['color-primary-400']}
    />
  );
  return (
    <>
      <View style={{flex: 1, backgroundColor: Colors.WHITE}}>
        {loadingSave || loadingReq ? (
          <View
            style={{flex: 1, alignContent: 'center', justifyContent: 'center'}}>
            <ActivityIndicator size={'large'} color={Colors.PRIMARY.red} />
          </View>
        ) : (
          <TabView
            navigationState={{index, routes}}
            renderScene={renderScene}
            onIndexChange={setIndex}
            renderTabBar={renderTabBar}
          />
        )}
      </View>
      {!route?.params?.isReadOnly && (
        <View
          style={{
            ...styles.rowForm,
            gap: Mixins.scaleSize(10),
            paddingHorizontal: Mixins.scaleSize(14),
            paddingVertical: Mixins.scaleSize(10),
          }}>
          <TouchableOpacity
            onPress={() => onSave()}
            disabled={loadingSave}
            style={{
              ...styles.buttonAction,
              backgroundColor: theme['color-primary-500'],
            }}>
            <Icon
              fill={Colors.WHITE}
              name="save-outline"
              style={{
                width: Mixins.scaleSize(20),
                height: Mixins.scaleSize(20),
              }}
            />
            <Text
              style={{
                ...styles.labelAction,
                color: Colors.WHITE,
              }}>
              Save
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            onPress={() => onRequest()}
            disabled={disabledRequest || loadingReq}
            style={{
              ...styles.buttonAction,
              backgroundColor: theme['color-primary-500'],
              opacity: disabledRequest ? 0.6 : 1,
            }}>
            <Icon
              fill={Colors.WHITE}
              name="save-outline"
              style={{
                width: Mixins.scaleSize(20),
                height: Mixins.scaleSize(20),
              }}
            />
            <Text
              style={{
                ...styles.labelAction,
                color: Colors.WHITE,
              }}>
              Request
            </Text>
          </TouchableOpacity>
        </View>
      )}
    </>
  );
};

export default container(RoutineAssetDetails, false);

const styles = StyleSheet.create({
  tabBarStyle: {
    backgroundColor: 'transparant',
    borderBottomColor: Colors.SHADES.gray[200],
    borderBottomWidth: 4,
    paddingBottom: 10,
    marginTop: Mixins.scaleSize(10),
  },
  tabbar: {
    backgroundColor: Colors.SHADES.gray[100],
  },
  tab: {
    width: Mixins.WINDOW_WIDTH / 2,
  },
  indicator: {
    backgroundColor: theme['color-primary-400'],
  },
  label: {
    ...Typography.textLg,
    textTransform: 'none',
    paddingHorizontal: Mixins.scaleSize(20),
    color: Colors.BLACK,
  },
  rowForm: {
    flexDirection: 'row',
  },
  buttonAction: {
    flex: 1,
    flexDirection: 'row',
    padding: Mixins.scaleSize(10),
    borderRadius: Mixins.scaleSize(5),
    justifyContent: 'center',
    alignItems: 'center',
    gap: Mixins.scaleSize(10),
  },
  labelAction: {
    ...Typography.textMdPlusSemiBold,
  },
});
